package consumer

import (
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"log/slog"
	"strings"
)

const (
	ErrInvalidMessageFormat = "invalid message format"
)

/*
	type Storage interface {
		SavePost(ctx context.Context, post *kafkav1.PostCreateEvent) error
		DeletePost(ctx context.Context, postID *kafkav1.PostDeletedEvent) error
		AddLike(ctx context.Context, like *kafkav1.PostLikedEvent) error
		AddComment(ctx context.Context, comment *kafkav1.CommentAddedEvent) error
	}
*/
const (
	// heartbeatTimeout is the Kafka consumer heartbeat timeout
	heartbeatTimeout = 7000 // ms

	// noTimeout is the Kafka consumer no timeout, bad practice, better to use a timeout
	noTimeout = -1 // no timeout
)

type Handler interface {
	HandleMessage(message []byte, offset kafka.Offset) error
}

// Consumer is a struct that represents a Kafka consumer
type Consumer struct {
	log      *slog.Logger
	consumer *kafka.Consumer
	handler  Handler
	stop     bool
}

// NewConsumer creates a new Kafka consumer instance
func NewConsumer(handler Handler, address []string, consumerGroup, topic string, log *slog.Logger) (*Consumer, error) {
	cfg := initConfig(address, consumerGroup)

	c, err := kafka.NewConsumer(&cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create consumer: %w", err)
	}

	err = c.Subscribe(topic, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to subscribe to topics: %w", err)
	}

	return &Consumer{
		handler:  handler,
		log:      log,
		consumer: c,
	}, nil
}

func (c *Consumer) Start() {
	const op = "consumer.Start"
	log := c.log.With("op", op)
	for {
		if c.stop {
			log.Info("stopping consumer")
			break
		}
		log.Info("waiting for messages")

		kafkaMsg, err := c.consumer.ReadMessage(noTimeout)
		if err != nil {
			log.Error("failed to read message", slog.String("error", err.Error()))
		}
		if kafkaMsg == nil {
			continue
		}
		log.Info(
			"received message",
			slog.String("topic", *kafkaMsg.TopicPartition.Topic),
			slog.String("message", string(kafkaMsg.Value)),
		)
		if err := c.handler.HandleMessage(kafkaMsg.Value, kafkaMsg.TopicPartition.Offset); err != nil {
			log.Error("failed to handle message", slog.String("error", err.Error()))
			continue
		}
		if _, err := c.consumer.StoreMessage(kafkaMsg); err != nil {
			log.Error("failed to store message", slog.String("error", err.Error()))
			continue
		}

	}
}
func (c *Consumer) Close() error {
	c.stop = true
	if _, err := c.consumer.Commit(); err != nil {
		c.log.Error("failed to close consumer", slog.String("error", err.Error()))
		return fmt.Errorf("failed to close consumer: %w", err)
	}
	return c.consumer.Close()
}

// initConfig initializes the Kafka producer configuration
func initConfig(address []string, consumerGroup string) kafka.ConfigMap {
	config := kafka.ConfigMap{
		"bootstrap.servers":        strings.Join(address, ","),
		"group.id":                 consumerGroup,
		"session.timeout.ms":       heartbeatTimeout,
		"auto.offset.reset":        "earliest",
		"enable.auto.offset.store": false,
		"enable.auto.commit":       false,
	}

	return config
}

// initTopics initializes the Kafka topics
func initTopics() []string {
	topicMap := []string{
		"post_created.events",
		"post_deleted.events",
		"interaction.events",
		"comment.events",
	}

	return topicMap
}
