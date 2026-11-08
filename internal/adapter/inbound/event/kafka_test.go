package event

import (
	"context"
	"log"
	"testing"
)

func TestPublish(t *testing.T) {
	adapter := NewKafkaAdapter()
	defer adapter.Close()

	err := adapter.Produce(
		context.Background(),
		Orders,
		[]byte("order-123"),
		[]byte("OK"),
	)
	if err != nil {
		log.Fatal(err)
	}
}
