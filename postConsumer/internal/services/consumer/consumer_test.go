package consumer

import (
	"context"
	kafkav1 "github.com/NorthDice/DeepLink/protos/gen/go/kafka"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
	"log/slog"
	"testing"
	"time"
)

type MockStorage struct {
	mock.Mock
}

func (m *MockStorage) SavePost(ctx context.Context, post *kafkav1.PostCreateEvent) error {
	args := m.Called(ctx, post)
	return args.Error(0)
}

func (m *MockStorage) DeletePost(ctx context.Context, postID *kafkav1.PostDeletedEvent) error {
	args := m.Called(ctx, postID)
	return args.Error(0)
}

func (m *MockStorage) AddLike(ctx context.Context, like *kafkav1.PostLikedEvent) error {
	args := m.Called(ctx, like)
	return args.Error(0)
}

func (m *MockStorage) AddComment(ctx context.Context, comment *kafkav1.CommentAddedEvent) error {
	args := m.Called(ctx, comment)
	return args.Error(0)
}

func TestKafkaConsumer_handlePostCreated(t *testing.T) {
	mockStorage := new(MockStorage)
	logger := slog.New(slog.NewTextHandler(nil, nil))

	kc := &KafkaConsumer{
		Log:      logger,
		Storage:  mockStorage,
		handlers: make(map[string]MessageHandler),
	}

	kc.handlers["post_created.events"] = kc.handlePostCreated

	event := &kafkav1.PostCreateEvent{
		PostId:    1,
		UserId:    1,
		Content:   "Test Content",
		MediaUrl:  []string{"http://example.com/image.jpg"},
		CreatedAt: timestamppb.New(time.Now().UTC()),
	}

	eventBytes, err := proto.Marshal(event)
	assert.NoError(t, err)

	mockStorage.On("SavePost", mock.Anything, mock.MatchedBy(func(p *kafkav1.PostCreateEvent) bool {
		return p.Content == "Test Content"
	})).Return(nil)

	err = kc.handlers["post_created.events"](context.Background(), eventBytes)
	assert.NoError(t, err)
	
	mockStorage.AssertExpectations(t)
}
