package consumer

import (
	"context"
	"errors"
	"fmt"
	"github.com/IBM/sarama"
	kafkav1 "github.com/NorthDice/DeepLink/protos/gen/go/kafka"
	"google.golang.org/protobuf/proto"
	"log/slog"
	"time"
)

const (
	rebalancedTimeout = 30 * time.Second
	processTimeout    = 10 * time.Second
)

const (
	ErrInvalidMessageFormat = "invalid message format"
)

type Storage interface {
	SavePost(ctx context.Context, post *kafkav1.PostCreateEvent) error
	DeletePost(ctx context.Context, postID *kafkav1.PostDeletedEvent) error
	AddLike(ctx context.Context, like *kafkav1.PostLikedEvent) error
	AddComment(ctx context.Context, comment *kafkav1.CommentAddedEvent) error
}

type KafkaConsumer struct {
	Log           *slog.Logger
	ConsumerGroup sarama.ConsumerGroup
	Storage       Storage
	topics        []string
	handlers      map[string]MessageHandler
}

type MessageHandler func(ctx context.Context, data []byte) error

func New(
	log *slog.Logger,
	brokers []string,
	groupID string,
	storage Storage,
	topics []string,
) (*KafkaConsumer, error) {
	config := sarama.NewConfig()
	config.Consumer.Group.Rebalance.Timeout = rebalancedTimeout
	config.Consumer.Offsets.Initial = sarama.OffsetOldest
	config.Consumer.Offsets.AutoCommit.Enable = true
	config.Consumer.Offsets.AutoCommit.Interval = 1 * time.Second

	consumerGroup, err := sarama.NewConsumerGroup(brokers, groupID, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create consumer group: %w", err)
	}

	kc := &KafkaConsumer{
		Log:           log,
		ConsumerGroup: consumerGroup,
		Storage:       storage,
		topics:        topics,
	}

	kc.handlers = kc.initHandlers()

	return kc, nil
}

func (kc *KafkaConsumer) initHandlers() map[string]MessageHandler {
	return map[string]MessageHandler{
		"post_created.events":  kc.handlePostCreated,
		"post_deleted.events":  kc.handlePostDeleted,
		"post_liked.events":    kc.handlePostLiked,
		"comment_added.events": kc.handleCommentAdded,
	}
}

func (kc *KafkaConsumer) handlePostCreated(ctx context.Context, data []byte) error {
	var event kafkav1.PostCreateEvent
	if err := proto.Unmarshal(data, &event); err != nil {
		return fmt.Errorf("failed to unmarshal post event: %w", err)
	}
	return kc.Storage.SavePost(ctx, &event)
}

func (kc *KafkaConsumer) handlePostDeleted(ctx context.Context, data []byte) error {
	var event kafkav1.PostDeletedEvent
	if err := proto.Unmarshal(data, &event); err != nil {
		return fmt.Errorf("failed to unmarshal post deleted event: %w", err)
	}
	return kc.Storage.DeletePost(ctx, &event)
}

func (kc *KafkaConsumer) handlePostLiked(ctx context.Context, data []byte) error {
	var event kafkav1.PostLikedEvent
	if err := proto.Unmarshal(data, &event); err != nil {
		return fmt.Errorf("failed to unmarshal post liked event: %w", err)
	}
	return kc.Storage.AddLike(ctx, &event)
}

func (kc *KafkaConsumer) handleCommentAdded(ctx context.Context, data []byte) error {
	var event kafkav1.CommentAddedEvent
	if err := proto.Unmarshal(data, &event); err != nil {
		return fmt.Errorf("failed to unmarshal comment added event: %w", err)
	}
	return kc.Storage.AddComment(ctx, &event)
}

func (kc *KafkaConsumer) Run(ctx context.Context) error {
	const op = "consumer.Run"
	log := kc.Log.With(slog.String("op", op))

	log.Info("starting consumer", slog.Any("topics", kc.topics))

	for {
		select {
		case <-ctx.Done():
			log.Info("shutting down by context")
			return nil
		default:
			err := kc.ConsumerGroup.Consume(ctx, kc.topics, kc)
			if err != nil {
				if errors.Is(err, sarama.ErrClosedConsumerGroup) {
					return nil
				}
				log.Error("consumption error", slog.String("error", err.Error()))
				time.Sleep(5 * time.Second)
			}
		}
	}
}

func (kc *KafkaConsumer) Close() error {
	const op = "KafkaConsumer.Close"
	log := kc.Log.With(slog.String("op", op))

	log.Info("closing consumer")
	if err := kc.ConsumerGroup.Close(); err != nil {
		log.Error("failed to close consumer", slog.String("error", err.Error()))
		return fmt.Errorf("%s: %w", op, err)
	}

	log.Info("consumer closed successfully")
	return nil
}

func (kc *KafkaConsumer) Setup(sarama.ConsumerGroupSession) error {
	kc.Log.Info("consumer group session started")
	return nil
}

func (kc *KafkaConsumer) Cleanup(sarama.ConsumerGroupSession) error {
	kc.Log.Info("consumer group session ended")
	return nil
}

func (kc *KafkaConsumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	const op = "services.ConsumeClaim"
	log := kc.Log.With(
		slog.String("op", op),
		slog.String("topic", claim.Topic()),
		slog.Int("partition", int(claim.Partition())),
	)

	log.Info("starting consuming messages")

	for msg := range claim.Messages() {
		select {
		case <-session.Context().Done():
			log.Warn("shutting down by context")
			return nil
		default:
		}

		log.Debug("message received",
			slog.Int64("offset", msg.Offset),
			slog.Time("timestamp", msg.Timestamp),
		)

		if err := kc.ProcessMessage(session.Context(), msg); err != nil {
			log.Error("failed to process message",
				slog.Int64("offset", msg.Offset),
				slog.String("error", err.Error()))
			continue
		}

		session.MarkMessage(msg, "")
		log.Debug("message processed successfully",
			slog.Int64("offset", msg.Offset))
	}
	return nil
}

func (kc *KafkaConsumer) ProcessMessage(ctx context.Context, msg *sarama.ConsumerMessage) error {
	const op = "services.processMessage"
	log := kc.Log.With(slog.String("op", op))

	ctx, cancel := context.WithTimeout(ctx, processTimeout)
	defer cancel()

	handler, exists := kc.handlers[msg.Topic]
	if !exists {
		return fmt.Errorf("%s: unknown topic %s", op, msg.Topic)
	}

	if err := handler(ctx, msg.Value); err != nil {
		log.Error("failed to process message",
			slog.String("topic", msg.Topic),
			slog.String("error", err.Error()))
		return fmt.Errorf("%s: %v: %w", op, ErrInvalidMessageFormat, err)
	}

	return nil
}
