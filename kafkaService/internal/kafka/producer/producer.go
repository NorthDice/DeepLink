package producer

import (
	"context"
	"fmt"
	"github.com/IBM/sarama"
	kafkav1 "github.com/NorthDice/DeepLink/protos/gen/go/kafka"
	"github.com/gogo/protobuf/proto"
	"log/slog"
	"time"
)

const (
	maxMessageSize = 1024 * 1024
)

type KafkaProducer struct {
	syncProducer sarama.SyncProducer
	topicConfig  map[string]string
}

func New(brokerList []string, log *slog.Logger) (*KafkaProducer, error) {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.MaxMessageBytes = maxMessageSize
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5

	producer, err := sarama.NewSyncProducer(brokerList, config)
	if err != nil {
		return nil, fmt.Errorf("create kafka producer fail, %s", err.Error())
	}

	return &KafkaProducer{
		syncProducer: producer,
		topicConfig: map[string]string{
			"post_created":  "post.events",
			"post_deleted":  "post.events",
			"post_liked":    "interaction.events",
			"comment_added": "comment.events",
		},
	}, nil
}

func (p *KafkaProducer) Close() error {
	return p.syncProducer.Close()
}

func (p *KafkaProducer) produce(ctx context.Context, topic string, msg proto.Message) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	if err := validateMessageSize(msg); err != nil {
		return err
	}

	value, err := proto.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal message fail, %s", err.Error())
	}

	_, _, err = p.syncProducer.SendMessage(&sarama.ProducerMessage{
		Topic:     p.topicConfig[topic],
		Value:     sarama.ByteEncoder(value),
		Timestamp: time.Now(),
	})

	if err != nil {
		return fmt.Errorf("send message fail, %s", err.Error())
	}
	return nil
}

func (p *KafkaProducer) ProducePostCreated(ctx context.Context, event *kafkav1.PostCreateEvent) error {
	const op = "KafkaProducer.ProducePostCreated"

	if event == nil {
		return fmt.Errorf("%s: event is nil", op)
	}
	return p.produce(ctx, "post_created", event)
}

func (p *KafkaProducer) ProducePostDeleted(ctx context.Context, event *kafkav1.PostCreateEvent) error {
	const op = "KafkaProducer.ProducePostDeleted"

	if event == nil {
		return fmt.Errorf("%s: event is nil", op)
	}

	return p.produce(ctx, "post_deleted", event)

}

func (p *KafkaProducer) ProducePostLiked(ctx context.Context, event *kafkav1.PostLikedEvent) error {
	const op = "KafkaProducer.ProducePostLiked"
	if event == nil {
		return fmt.Errorf("%s: event is nil", op)
	}
	if event.PostId <= 0 || event.UserId <= 0 {
		return fmt.Errorf("%s: post id or user id should be greater than 0", op)
	}
	return p.produce(ctx, "post_liked", event)
}

func (p *KafkaProducer) ProduceCommentAdded(ctx context.Context, event *kafkav1.CommentAddedEvent) error {
	const op = "KafkaProducer.ProduceCommentAdded"

	if event == nil {
		return fmt.Errorf("%s: event is nil", op)
	}

	if event.PostId <= 0 || event.UserId <= 0 {
		return fmt.Errorf("%s: post id or user id should be greater than 0", op)
	}
	return p.produce(ctx, "comment_added", event)
}

func validateMessageSize(msg proto.Message) error {

	if size := proto.Size(msg); size > maxMessageSize {
		return fmt.Errorf("message to large (%d > %d)", size, maxMessageSize)
	}

	return nil
}
