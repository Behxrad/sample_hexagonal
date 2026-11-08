package event

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	inventory "192.168.100.6/coreinno/proto_common.git/gen/sample/inventory/v1"
	inbound_headers "192.168.100.6/coreinno/proto_common.git/model/common/v1"
	"github.com/Behxrad/sample_hexagonal/configs"
	"github.com/Behxrad/sample_hexagonal/internal/port/inbound/api"
	"github.com/segmentio/kafka-go"
	"google.golang.org/protobuf/proto"
)

const (
	defaultPartitions        = 10
	defaultReplicationFactor = 1
	productAddGroupID        = "product-processors"
)

type topicName string

var (
	ProductAdd     = makeTopicName("product-add")
	ProductCreated = makeTopicName("product-created")
)

type KafkaAdapter struct {
	brokers []string
	svc     api.ProductService

	mu      sync.RWMutex
	writers map[topicName]*kafka.Writer
	cancel  context.CancelFunc
	wg      sync.WaitGroup
	closed  bool
}

func makeTopicName(name string) topicName {
	return topicName(fmt.Sprintf("%s-%s", configs.ApplicationConfigs.ServiceName, name))
}

func topicConfigs() []kafka.TopicConfig {
	topics := []topicName{ProductAdd, ProductCreated}
	cfgs := make([]kafka.TopicConfig, 0, len(topics))
	for _, t := range topics {
		cfgs = append(cfgs, kafka.TopicConfig{
			Topic:             string(t),
			NumPartitions:     defaultPartitions,
			ReplicationFactor: defaultReplicationFactor,
		})
	}
	return cfgs
}

func NewKafkaAdapter(svc api.ProductService) *KafkaAdapter {
	if svc == nil {
		log.Println("product service is required")
	}

	addr := fmt.Sprintf("%s:%d",
		configs.ApplicationConfigs.EventBrokerHost,
		configs.ApplicationConfigs.EventBrokerPort,
	)

	return &KafkaAdapter{
		brokers: []string{addr},
		writers: make(map[topicName]*kafka.Writer),
		svc:     svc,
	}
}

func (k *KafkaAdapter) EnsureTopics(ctx context.Context) error {
	timeout, err := time.ParseDuration(configs.ApplicationConfigs.EventBrokerTimeout)
	if err != nil {
		return fmt.Errorf("parse kafka timeout: %w", err)
	}

	dialCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	conn, err := kafka.DialContext(dialCtx, "tcp", k.brokers[0])
	if err != nil {
		return fmt.Errorf("dial kafka: %w", err)
	}
	defer conn.Close()

	if err := conn.CreateTopics(topicConfigs()...); err != nil {
		return fmt.Errorf("create topics: %w", err)
	}
	return nil
}

func (k *KafkaAdapter) getWriter(topic topicName) (*kafka.Writer, error) {
	k.mu.Lock()
	defer k.mu.Unlock()

	if k.closed {
		return nil, errors.New("kafka adapter is closed")
	}

	if w, ok := k.writers[topic]; ok {
		return w, nil
	}

	w := &kafka.Writer{
		Addr:         kafka.TCP(k.brokers...),
		Topic:        string(topic),
		Balancer:     &kafka.Hash{},
		Async:        false,
		RequiredAcks: kafka.RequireAll,
	}
	k.writers[topic] = w
	return w, nil
}

func (k *KafkaAdapter) Produce(ctx context.Context, topic topicName, key, value []byte) error {
	w, err := k.getWriter(topic)
	if err != nil {
		return err
	}

	return w.WriteMessages(ctx, kafka.Message{
		Key:   key,
		Value: value,
		Time:  time.Now(),
	})
}

func (k *KafkaAdapter) Consume(
	ctx context.Context,
	topic topicName,
	groupID string,
	handler func(context.Context, kafka.Message) error,
) error {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  k.brokers,
		Topic:    string(topic),
		GroupID:  groupID,
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
		MaxWait:  time.Second,
	})
	defer func() {
		if err := r.Close(); err != nil {
			log.Printf("close kafka reader: \ntopic: %s\nerr: %s", topic, err.Error())
		}
	}()

	for {
		msg, err := r.FetchMessage(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return nil
			}
			return fmt.Errorf("fetch message from %s: %w", topic, err)
		}

		if err := handler(ctx, msg); err != nil {
			log.Printf("handle message from %s, err: %s", topic, err.Error())
			continue
		}

		if err := r.CommitMessages(ctx, msg); err != nil {
			return fmt.Errorf("commit message from %s: %w", topic, err)
		}
	}
}

func (k *KafkaAdapter) Start(ctx context.Context) error {
	k.mu.Lock()
	defer k.mu.Unlock()

	if k.closed {
		return errors.New("kafka adapter is closed")
	}
	if k.cancel != nil {
		return errors.New("kafka adapter already started")
	}

	readerCtx, cancel := context.WithCancel(ctx)
	k.cancel = cancel

	k.wg.Add(1)
	go func() {
		defer k.wg.Done()

		err := k.Consume(readerCtx, ProductAdd, productAddGroupID, k.handleProductAdd)
		if err != nil && readerCtx.Err() == nil {
			log.Println("kafka consumer stopped unexpectedly", "error", err)
			cancel()
		}
	}()

	return nil
}

func (k *KafkaAdapter) Close() error {
	k.mu.Lock()
	if k.closed {
		k.mu.Unlock()
		return nil
	}
	k.closed = true
	cancel := k.cancel
	k.mu.Unlock()

	if cancel != nil {
		cancel()
	}

	k.wg.Wait()

	k.mu.Lock()
	defer k.mu.Unlock()

	var closeErr error
	for topic, w := range k.writers {
		if err := w.Close(); err != nil {
			closeErr = errors.Join(closeErr, fmt.Errorf("close writer %s: %w", topic, err))
		}
	}
	k.writers = nil
	return closeErr
}

func (k *KafkaAdapter) handleProductAdd(ctx context.Context, msg kafka.Message) error {
	var cmd inventory.CreateProductCommand
	if err := proto.Unmarshal(msg.Value, &cmd); err != nil {
		return fmt.Errorf("unmarshal CreateProductCommand: %w", err)
	}

	result, err := k.svc.AddProduct(ctx, api.AddProductCommand{
		MetaData: inbound_headers.FromProto(cmd.Metadata),
		Product: api.Product{
			Count: uint16(cmd.Count), //nolint:gosec // domain constraint
			Price: uint32(cmd.Price), //nolint:gosec
			Name:  cmd.Name,
		},
	})
	if err != nil {
		return fmt.Errorf("add product: %w", err)
	}

	if err := k.publishProductCreated(ctx, result); err != nil {
		return fmt.Errorf("publish product created: %w", err)
	}
	return nil
}

func (k *KafkaAdapter) publishProductCreated(ctx context.Context, result api.AddProductResult) error {
	payload, err := proto.Marshal(&inventory.CreateProductResult{
		Metadata: result.MetaData.ToProto(),
		Id:       result.ID,
	})
	if err != nil {
		return fmt.Errorf("marshal CreateProductResult: %w", err)
	}

	return k.Produce(ctx, ProductCreated, []byte(result.ID), payload)
}
