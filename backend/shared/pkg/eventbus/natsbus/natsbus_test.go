package natsbus

import (
	"context"
	"os"
	"testing"
	"time"

	"sen1or/letslive/shared/pkg/eventbus"
	"sen1or/letslive/shared/pkg/logger"
)

func TestMain(m *testing.M) {
	// every logger.* call assumes main() has run logger.Init() first, same
	// as every service's cmd/main.go
	logger.Init(logger.Debug)
	os.Exit(m.Run())
}

func testNatsURL() string {
	if u := os.Getenv("NATS_TEST_URL"); u != "" {
		return u
	}
	return "nats://127.0.0.1:4222"
}

// TestProducerConsumerRoundTrip exercises Admin, Producer, and Consumer
// against a real NATS server. Skips if none is reachable — set
// NATS_TEST_URL to point at one, or run `docker run --rm -p 4222:4222
// nats:2-alpine -js`.
func TestProducerConsumerRoundTrip(t *testing.T) {
	url := testNatsURL()

	connectCtx, cancelConnect := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelConnect()

	admin, err := NewAdmin(connectCtx, url)
	if err != nil {
		t.Skipf("nats not reachable at %s, skipping integration test: %v", url, err)
	}
	defer admin.Close()

	topic := "letslive.test_roundtrip"
	if err := admin.EnsureTopics(connectCtx, []eventbus.TopicConfig{{Name: topic}}); err != nil {
		t.Fatalf("EnsureTopics failed: %v", err)
	}

	producer, err := NewProducer(connectCtx, url)
	if err != nil {
		t.Fatalf("NewProducer failed: %v", err)
	}
	defer producer.Close()

	consumer, err := NewConsumer(connectCtx, url, "test-consumer-group")
	if err != nil {
		t.Fatalf("NewConsumer failed: %v", err)
	}
	defer consumer.Close()

	consumeCtx, consumeCancel := context.WithCancel(context.Background())
	defer consumeCancel()

	received := make(chan eventbus.Event, 1)
	go func() {
		_ = consumer.Subscribe(consumeCtx, []string{topic}, func(_ context.Context, event eventbus.Event) error {
			received <- event
			return nil
		})
	}()

	// give the durable consumer a moment to attach before publishing
	time.Sleep(300 * time.Millisecond)

	event, err := eventbus.NewEvent("test.roundtrip", "natsbus-test", map[string]string{"hello": "world"})
	if err != nil {
		t.Fatalf("NewEvent failed: %v", err)
	}

	if err := producer.Publish(connectCtx, topic, "test-key", event); err != nil {
		t.Fatalf("Publish failed: %v", err)
	}

	select {
	case got := <-received:
		if got.ID != event.ID {
			t.Fatalf("got event id %s, want %s", got.ID, event.ID)
		}
		if got.Type != event.Type {
			t.Fatalf("got event type %s, want %s", got.Type, event.Type)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for event to be consumed")
	}
}
