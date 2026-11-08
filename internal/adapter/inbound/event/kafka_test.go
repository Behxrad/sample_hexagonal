package event

import (
	"context"
	"log"
	"sync"
	"testing"
	"time"

	common "192.168.100.6/coreinno/proto_common.git/gen/common/v1"
	inventory "192.168.100.6/coreinno/proto_common.git/gen/sample/inventory/v1"
	"github.com/Behxrad/sample_hexagonal/internal/adapter/outbound/repository"
	"github.com/Behxrad/sample_hexagonal/internal/service"
	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
	"google.golang.org/protobuf/proto"
)

func TestProduce(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	productRepo := repository.NewMongoProductRepository(ctx)
	productService := service.NewProductService(productRepo)

	adapter := NewKafkaAdapter(productService)
	defer func() {
		if err := adapter.Close(); err != nil {
			t.Errorf("Close: %v", err)
		}
	}()

	if err := adapter.EnsureTopics(ctx); err != nil {
		t.Fatalf("EnsureTopics: %v", err)
	}

	reqID := uuid.NewString()
	cmd := &inventory.CreateProductCommand{
		Metadata: &common.RequestMetadata{
			RequestTimeUnixMs: time.Now().UnixMilli(),
			RequestId:         reqID,
		},
		Count: 10,
		Price: 2000,
		Name:  "pencil10",
	}

	payload, err := proto.Marshal(cmd)
	if err != nil {
		t.Fatalf("marshal command: %v", err)
	}

	if err := adapter.Produce(ctx, ProductAdd, []byte(reqID), payload); err != nil {
		t.Fatalf("Produce: %v", err)
	}

	log.Printf("produced CreateProductCommand with requestId=%s", reqID)
}

func TestConsume(t *testing.T) {
	ctx := context.Background()

	productRepo := repository.NewMongoProductRepository(ctx)
	productService := service.NewProductService(productRepo)

	adapter := NewKafkaAdapter(productService)
	defer func() {
		if err := adapter.Close(); err != nil {
			t.Errorf("Close: %v", err)
		}
	}()

	if err := adapter.EnsureTopics(ctx); err != nil {
		t.Fatalf("EnsureTopics: %v", err)
	}

	var (
		wg     sync.WaitGroup
		result *inventory.CreateProductResult
		mu     sync.Mutex
	)

	wg.Go(func() {
		err := adapter.Consume(ctx, ProductCreated, "test-product-created-group",
			func(ctx context.Context, msg kafka.Message) error {
				var res inventory.CreateProductResult
				if err := proto.Unmarshal(msg.Value, &res); err != nil {
					return err
				}

				mu.Lock()
				result = &res
				mu.Unlock()

				log.Printf("consumed ProductCreated: %s", res.String())
				return nil
			},
		)
		if err != nil && ctx.Err() == nil {
			t.Errorf("Consume: %v", err)
		}
	})

	// Give consumer time to start
	time.Sleep(500 * time.Millisecond)

	// Produce a message so the pipeline has something to process
	reqID := uuid.NewString()
	cmd := &inventory.CreateProductCommand{
		Metadata: &common.RequestMetadata{
			RequestTimeUnixMs: time.Now().UnixMilli(),
			RequestId:         reqID,
		},
		Count: 10,
		Price: 2000,
		Name:  "pencil10",
	}

	payload, err := proto.Marshal(cmd)
	if err != nil {
		t.Fatalf("marshal command: %v", err)
	}

	if err := adapter.Produce(ctx, ProductAdd, []byte(reqID), payload); err != nil {
		t.Fatalf("Produce: %v", err)
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-ctx.Done():
		t.Fatal("timed out waiting for ProductCreated event")
	}

	mu.Lock()
	defer mu.Unlock()

	if result == nil {
		t.Fatal("expected ProductCreated result, got nil")
	}

	if result.Id == "" {
		t.Error("expected non-empty product id")
	}

	log.Printf("final result id=%s requestId=%s", result.Id, result.Metadata.Description)
}
