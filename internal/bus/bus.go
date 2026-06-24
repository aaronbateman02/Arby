package bus

import (
	"encoding/json"
	"fmt"
	"sync"
)

type Message struct {
	Topic   string
	Payload []byte
}

type Bus struct {
	mu          sync.RWMutex
	subscribers map[string]map[int]chan Message
	nextID      int
}

func New(channelSize int) *Bus {
	return &Bus{
		subscribers: make(map[string]map[int]chan Message),
	}
}

func (b *Bus) Subscribe(topic string) (chan Message, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.subscribers[topic] == nil {
		b.subscribers[topic] = make(map[int]chan Message)
	}

	b.nextID++
	ch := make(chan Message, 100)
	b.subscribers[topic][b.nextID] = ch

	return ch, nil
}

func (b *Bus) Unsubscribe(topic string, ch chan Message) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if subs, ok := b.subscribers[topic]; ok {
		for id, subCh := range subs {
			if subCh == ch {
				close(subCh)
				delete(subs, id)
				return
			}
		}
	}
}

func (b *Bus) Publish(topic string, payload []byte) {
	b.mu.RLock()
	subs := b.subscribers[topic]
	b.mu.RUnlock()

	if subs == nil {
		return
	}

	msg := Message{Topic: topic, Payload: payload}

	b.mu.RLock()
	for _, ch := range subs {
		select {
		case ch <- msg:
		default:
		}
	}
	b.mu.RUnlock()
}

func (b *Bus) PublishTyped(topic string, v interface{}) error {
	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("marshal event %s: %w", topic, err)
	}
	b.Publish(topic, data)
	return nil
}
