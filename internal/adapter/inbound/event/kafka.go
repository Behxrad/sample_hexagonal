package event

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/Behxrad/sample_hexagonal/configs"
	"github.com/segmentio/kafka-go"
)

const (
	partitionsNum     = 10
	replicationFactor = 1
)

type TopicName string

var (
	Orders        = createTopicName("orders")
	Payments      = createTopicName("payments")
	Notifications = createTopicName("notifications")
)

type KafkaAdapter struct {
	brokers []string

	mu      sync.Mutex
	writers map[TopicName]*kafka.Writer
}

func createTopicName(name string) TopicName {
	return TopicName(fmt.Sprintf(
		"%s-%s",
		configs.ApplicationConfigs.ServiceName,
		name,
	))
}

func getTopicConfigs() []kafka.TopicConfig {
	names := []TopicName{Orders, Payments, Notifications}
	confs := make([]kafka.TopicConfig, 0, len(names))

	for _, name := range names {
		confs = append(confs, kafka.TopicConfig{
			Topic:             string(name),
			NumPartitions:     partitionsNum,
			ReplicationFactor: replicationFactor,
		})
	}

	return confs
}

func NewKafkaAdapter() *KafkaAdapter {
	address := fmt.Sprintf(
		"%s:%d",
		configs.ApplicationConfigs.EventBrokerHost,
		configs.ApplicationConfigs.EventBrokerPort,
	)

	timeout, err := time.ParseDuration(
		configs.ApplicationConfigs.EventBrokerTimeout,
	)
	if err != nil {
		log.Println("parse Kafka timeout: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	conn, err := kafka.DialContext(ctx, "tcp", address)
	if err != nil {
		log.Println("connect to Kafka: %w", err)
	}

	if err := conn.CreateTopics(getTopicConfigs()...); err != nil {
		_ = conn.Close()
		log.Println("create Kafka topics: %w", err)
	}

	if err := conn.Close(); err != nil {
		log.Println("close Kafka admin connection: %w", err)
	}

	return &KafkaAdapter{
		brokers: []string{address},
		writers: make(map[TopicName]*kafka.Writer),
	}
}

func (k *KafkaAdapter) Close() error {
	k.mu.Lock()
	defer k.mu.Unlock()

	var closeErr error

	for topic, writer := range k.writers {
		if err := writer.Close(); err != nil {
			closeErr = errors.Join(
				closeErr,
				fmt.Errorf("close writer for topic %s: %w", topic, err),
			)
		}
	}

	return closeErr
}

func (k *KafkaAdapter) getWriter(topic TopicName) *kafka.Writer {
	k.mu.Lock()
	defer k.mu.Unlock()

	writer, exists := k.writers[topic]
	if exists {
		return writer
	}

	writer = &kafka.Writer{
		Addr:     kafka.TCP(k.brokers...),
		Topic:    string(topic),
		Balancer: &kafka.Hash{},
		Async:    false,
	}

	k.writers[topic] = writer

	return writer
}

func (k *KafkaAdapter) Produce(
	ctx context.Context,
	topic TopicName,
	key []byte,
	value []byte,
) error {
	writer := k.getWriter(topic)

	return writer.WriteMessages(ctx, kafka.Message{
		Key:   key,
		Value: value,
		Time:  time.Now(),
	})
}

func (k *KafkaAdapter) Consume(
	ctx context.Context,
	topic TopicName,
	groupID string,
	handler func(context.Context, kafka.Message) error,
) error {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  k.brokers,
		Topic:    string(topic),
		GroupID:  groupID,
		MinBytes: 10e3,
		MaxBytes: 10e6,
	})

	defer reader.Close()

	for {
		message, err := reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}

			return fmt.Errorf("fetch Kafka message: %w", err)
		}

		if err := handler(ctx, message); err != nil {
			return fmt.Errorf("process Kafka message: %w", err)
		}

		if err := reader.CommitMessages(ctx, message); err != nil {
			return fmt.Errorf("commit Kafka message: %w", err)
		}
	}
}

func (k *KafkaAdapter) Start(ctx context.Context) error {
	var mErr error
	go func() {
		err := k.Consume(ctx, Orders, "product-processors", processOrder)
		if err != nil {
			errors.Join(mErr, err)
		}
	}()
	go func() {
		err := k.Consume(ctx, Notifications, "product-processors", processOrder)
		if err != nil {
			errors.Join(mErr, err)
		}
	}()
	go func() {
		err := k.Consume(ctx, Payments, "product-processors", processOrder)
		if err != nil {
			errors.Join(mErr, err)
		}
	}()
	if mErr != nil {
		return mErr
	}
	return nil
}

func processOrder(ctx context.Context, message kafka.Message) error {
	log.Printf("process order message:\n%+v\n", message)
	log.Println(string(message.Value))
	return nil
}
