package mongodb

import (
	"context"
	"fmt"
	kafkav1 "github.com/NorthDice/DeepLink/protos/gen/go/kafka"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"log/slog"
	"time"
)

type MongoStorage struct {
	log          *slog.Logger
	client       *mongo.Client
	databaseName string
}

var (
	ErrDuplicateKey = "duplicate key"
	ErrNotFound     = "not found"
)

func New(
	uri string,
	databaseName string,
) (*MongoStorage, error) {
	const op = "storage.mongodb.New"

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := Connect(ctx, uri)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &MongoStorage{
		client:       client,
		databaseName: databaseName,
	}, nil
}

func (s *MongoStorage) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return s.client.Disconnect(ctx)
}
func Connect(ctx context.Context, uri string) (*mongo.Client, error) {

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}

	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		return nil, err
	}

	return client, nil
}

func (s *MongoStorage) SavePost(ctx context.Context, post *kafkav1.PostCreateEvent) error {
	const op = "storage.mongodb.SavePost"

	log := s.log.With(slog.String("op", op))

	collection := s.client.Database(s.databaseName).Collection("posts")

	doc := bson.M{
		"post_id":    post.PostId,
		"user_id":    post.UserId,
		"content":    post.Content,
		"media_urls": post.MediaUrl,
		"created_at": post.CreatedAt.AsTime(),
		"updated_at": time.Now(),
	}
	_, err := collection.InsertOne(ctx, doc)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			log.Warn("post already exists", slog.Int64("post_id", post.PostId))

			return fmt.Errorf("%s: %w", op, ErrDuplicateKey)
		}

		log.Error("failed to insert post", slog.String("error", err.Error()))
		return fmt.Errorf("%s: %w", op, err)
	}

	log.Info("post saved successfully", slog.Int64("post_id", post.PostId))
	return nil
}

func (s *MongoStorage) DeletePost(ctx context.Context, event *kafkav1.PostDeletedEvent) error {
	const op = "storage.mongodb.DeletePost"
	log := s.log.With(slog.String("op", op), slog.Int64("post_id", event.PostId))

	collection := s.client.Database(s.databaseName).Collection("posts")

	filter := bson.M{"post_id": event.PostId}
	result, err := collection.DeleteOne(ctx, filter)
	if err != nil {
		log.Error("failed to delete post", slog.String("error", err.Error()))
		return fmt.Errorf("%s: %w", op, err)
	}

	if result.DeletedCount == 0 {
		log.Warn("post not found", slog.Int64("post_id", event.PostId))
		return fmt.Errorf("%s: %w", op, ErrNotFound)
	}

	log.Info("post deleted", slog.Int64("post_id", event.PostId))
	return nil
}

func (s *MongoStorage) AddLike(ctx context.Context, event *kafkav1.PostLikedEvent) error {
	const op = "storage.mongodb.AddLike"
	log := s.log.With(
		slog.String("op", op),
		slog.Int64("post_id", event.PostId),
		slog.Int64("user_id", event.UserId),
	)

	collection := s.client.Database(s.databaseName).Collection("likes")

	doc := bson.M{
		"post_id":    event.PostId,
		"user_id":    event.UserId,
		"created_at": time.Now(),
	}

	_, err := collection.InsertOne(ctx, doc)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			log.Warn("Already liked")
			return nil
		}
		log.Error("failed to insert post", slog.String("error", err.Error()))
		return fmt.Errorf("%s: %w", op, err)

	}

	postCollection := s.client.Database(s.databaseName).Collection("posts")
	update := bson.M{"$inc": bson.M{"likes_count": 1}}
	_, err = postCollection.UpdateOne(ctx, bson.M{"post_id": event.PostId}, update)
	if err != nil {
		log.Error("failed to update likes", slog.String("error", err.Error()))
	}

	log.Info("post liked", slog.Int64("post_id", event.PostId))
	return nil
}
func (s *MongoStorage) AddComment(ctx context.Context, event *kafkav1.CommentAddedEvent) error {
	const op = "MongoStorage.AddComment"
	log := s.log.With(
		slog.String("op", op),
		slog.Int64("post_id", event.PostId),
		slog.Int64("user_id", event.UserId),
	)

	collection := s.client.Database(s.databaseName).Collection("comments")

	doc := bson.M{
		"post_id":    event.PostId,
		"user_id":    event.UserId,
		"content":    event.Content,
		"created_at": event.CreatedAt.AsTime(),
	}

	_, err := collection.InsertOne(ctx, doc)
	if err != nil {
		log.Error("failed to add event", slog.String("error", err.Error()))
		return fmt.Errorf("%s: %w", op, err)
	}
	
	postsCollection := s.client.Database(s.databaseName).Collection("posts")
	update := bson.M{"$inc": bson.M{"comments_count": 1}}
	_, err = postsCollection.UpdateOne(
		ctx,
		bson.M{"post_id": event.PostId},
		update,
	)
	if err != nil {
		log.Error("failed to update comments count", slog.String("error", err.Error()))
	}

	log.Info("event added successfully")
	return nil
}
