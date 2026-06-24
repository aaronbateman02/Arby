package bus_test

import (
	"testing"
	"time"

	"github.com/aaronbateman02/Arby/internal/bus"
)

func TestPublishSubscribe(t *testing.T) {
	b := bus.New(10)

	sub, err := b.Subscribe("test.event")
	if err != nil {
		t.Fatalf("subscribe failed: %v", err)
	}
	defer b.Unsubscribe("test.event", sub)

	payload := []byte("hello")
	b.Publish("test.event", payload)

	select {
	case msg := <-sub:
		if string(msg.Payload) != "hello" {
			t.Fatalf("expected 'hello', got '%s'", string(msg.Payload))
		}
		if msg.Topic != "test.event" {
			t.Fatalf("expected topic 'test.event', got '%s'", msg.Topic)
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for message")
	}
}

func TestUnsubscribe_StopsDelivery(t *testing.T) {
	b := bus.New(10)

	sub, _ := b.Subscribe("test.event")
	b.Unsubscribe("test.event", sub)

	b.Publish("test.event", []byte("should-not-receive"))

	select {
	case <-sub:
		t.Fatal("should not receive message after unsubscribe")
	case <-time.After(100 * time.Millisecond):
	}
}
