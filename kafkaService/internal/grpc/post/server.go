package post

import (
	"context"
	kafkav1 "github.com/NorthDice/DeepLink/protos/gen/go/kafka"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Producer interface {
	ProducePostCreated(ctx context.Context, event *kafkav1.PostCreateEvent) error
	ProducePostDeleted(ctx context.Context, event *kafkav1.PostDeletedEvent) error
	ProducePostLiked(ctx context.Context, event *kafkav1.PostLikedEvent) error
	ProduceCommentAdded(ctx context.Context, event *kafkav1.CommentAddedEvent) error
}

type serverAPI struct {
	kafkav1.UnimplementedKafkaServiceServer
	producer Producer
}

func Register(gRPCServer *grpc.Server) {
	kafkav1.RegisterKafkaServiceServer(gRPCServer, &serverAPI{})
}

func (s *serverAPI) CreatePost(ctx context.Context, req *kafkav1.PostCreateEvent) (*emptypb.Empty, error) {
	if err := s.producer.ProducePostCreated(ctx, req); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (s *serverAPI) DeletePost(ctx context.Context, req *kafkav1.PostDeletedEvent) (*emptypb.Empty, error) {
	if err := s.producer.ProducePostDeleted(ctx, req); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (s *serverAPI) LikePost(ctx context.Context, req *kafkav1.PostLikedEvent) (*emptypb.Empty, error) {
	if err := s.producer.ProducePostLiked(ctx, req); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *serverAPI) AddComment(ctx context.Context, req *kafkav1.CommentAddedEvent) (*emptypb.Empty, error) {
	if err := s.producer.ProduceCommentAdded(ctx, req); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}
