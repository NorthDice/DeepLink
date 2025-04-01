package consumer

import (
	"context"
	"errors"
	"fmt"
	"github.com/IBM/sarama"
	kafkav1 "github.com/NorthDice/DeepLink/protos/gen/go/kafka"
	"github.com/gogo/protobuf/proto"
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
	log           *slog.Logger
	consumerGroup sarama.ConsumerGroup
	storage       Storage
	topics        []string
}

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

	return &KafkaConsumer{
		log:           log,
		consumerGroup: consumerGroup,
		storage:       storage,
		topics:        topics,
	}, nil
}

func (kc *KafkaConsumer) Run(ctx context.Context) error {
	const op = "consumer.Run"

	log := kc.log.With(slog.String("op", op))

	log.Info("starting consumer", slog.Any("topics", kc.topics))

	for {
		select {
		case <-ctx.Done():
			log.Info("shutting down by context")
			return nil
		default:
			err := kc.consumerGroup.Consume(ctx, kc.topics, kc)
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
	log := kc.log.With(slog.String("op", op))

	log.Info("closing consumer")
	if err := kc.consumerGroup.Close(); err != nil {
		log.Error("failed to close consumer", slog.String("error", err.Error()))
		return fmt.Errorf("%s: %w", op, err)
	}

	log.Info("consumer closed successfully")
	return nil
}

func (kc *KafkaConsumer) Setup(sarama.ConsumerGroupSession) error {
	kc.log.Info("consumer group session started")
	return nil
}

func (kc *KafkaConsumer) Cleanup(sarama.ConsumerGroupSession) error {
	kc.log.Info("consumer group session ended")
	return nil
}

func (kc *KafkaConsumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	const op = "kafka.ConsumeClaim"
	log := kc.log.With(slog.String("op", op),
		slog.String("topic", claim.Topic()),
		slog.Int("partition", int(claim.Partition())),
	)

	log.Info("starting comsuming messages")

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

		if err := kc.processMessage(session.Context(), msg); err != nil {
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

func (kc *KafkaConsumer) processMessage(ctx context.Context, msg *sarama.ConsumerMessage) error {
	const op = "kafka.processMessage"
	log := kc.log.With(slog.String("op", op))

	ctx, cancel := context.WithTimeout(ctx, processTimeout)
	defer cancel()

	switch msg.Topic {
	case "post_created.events":
		var event kafkav1.PostCreateEvent
		if err := proto.Unmarshal(msg.Value, &event); err != nil {
			log.Error("failed to unmarshal post event", slog.String("error", err.Error()))

			return fmt.Errorf("%s: %w", op, ErrInvalidMessageFormat)
		}
		if err := kc.storage.SavePost(ctx, &event); err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}

	case "post_deleted.events":
		var event kafkav1.PostDeletedEvent
		if err := proto.Unmarshal(msg.Value, &event); err != nil {
			log.Error("failed to unmarshal post event", slog.String("error", err.Error()))

			return fmt.Errorf("%s: %w", op, ErrInvalidMessageFormat)
		}
		if err := kc.storage.DeletePost(ctx, &event); err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}

	case "post_liked.events":
		var event kafkav1.PostLikedEvent
		if err := proto.Unmarshal(msg.Value, &event); err != nil {
			log.Error("failed to unmarshal post event", slog.String("error", err.Error()))

			return fmt.Errorf("%s: %w", op, ErrInvalidMessageFormat)
		}
		if err := kc.storage.AddLike(ctx, &event); err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}

	case "comment_added.events":
		var event kafkav1.CommentAddedEvent
		if err := proto.Unmarshal(msg.Value, &event); err != nil {
			log.Error("failed to unmarshal comment added event", slog.String("error", err.Error()))

			return fmt.Errorf("%s: %w", op, ErrInvalidMessageFormat)
		}
		if err := kc.storage.AddComment(ctx, &event); err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}

	default:
		return fmt.Errorf("%s: unknown topic %s", op, msg.Topic)
	}

	return nil
}
