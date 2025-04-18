package producer

import (
	"errors"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"log/slog"
	"strings"
)

// flushTimeout is the timeout for flushing messages
const flushTimeout = 5000 //ms

// errUnknownType is returned when the event type is not recognized
var errUnknownType = errors.New("unknown type")

// Producer the structure for the Kafka producer
type Producer struct {
	log      *slog.Logger
	producer *kafka.Producer
	topicMap map[string]string
}

// NewProducer creates a new Kafka producer
func NewProducer(address []string, log *slog.Logger) (*Producer, error) {
	config := initConfig(address)
	topics := initTopics()

	p, err := kafka.NewProducer(&config)
	if err != nil {
		return nil, fmt.Errorf("failed to create producer: %w", err)
	}

	return &Producer{
		log:      log,
		producer: p,
		topicMap: topics,
	}, nil
}

// Produce sends a message to the appropriate Kafka topic based on the event type
func (p *Producer) Produce(message string, eventType string) error {
	topic, ok := p.topicMap[eventType]
	if !ok {
		return fmt.Errorf("%w: unknown event type: %s", errUnknownType, eventType)
	}

	kafkaMsg := kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &topic,
			Partition: kafka.PartitionAny,
		},
		Value: []byte(message),
		Key:   nil,
	}

	kafkaChan := make(chan kafka.Event)
	if err := p.producer.Produce(&kafkaMsg, kafkaChan); err != nil {
		return fmt.Errorf("failed to produce message: %w", err)
	}

	e := <-kafkaChan

	switch ev := e.(type) {
	case *kafka.Message:
		return nil
	case kafka.Error:
		return fmt.Errorf("kafka error: %s", ev)
	default:
		return errUnknownType
	}
}

func (p *Producer) Close() {
	p.producer.Flush(flushTimeout)
	p.producer.Close()
}

// initConfig initializes the Kafka producer configuration
func initConfig(address []string) kafka.ConfigMap {
	config := kafka.ConfigMap{
		"bootstrap.servers": strings.Join(address, ","),
	}

	return config
}

// initTopics initializes the Kafka topics
func initTopics() map[string]string {
	topicMap := map[string]string{
		"post_created":  "post_created.events",
		"post_deleted":  "post_deleted.events",
		"post_liked":    "interaction.events",
		"comment_added": "comment.events",
	}

	return topicMap
}
