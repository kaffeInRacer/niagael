package kafka

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/segmentio/kafka-go"
)

type Message = kafka.Message

type Watcher struct {
	writer   *kafka.Writer
	reader   *kafka.Reader
	callback func(string)
	cancel   context.CancelFunc
	mu       sync.Mutex
	closed   bool
}

func NewWatcher(brokers []string, topic, groupID string) *Watcher {
	return &Watcher{
		writer: NewWriter(brokers, topic),
		reader: NewReader(brokers, topic, groupID),
	}
}

// SetUpdateCallback starts watching the topic in the background and invokes
// the callback for every message received.
func (w *Watcher) SetUpdateCallback(callback func(string)) error {
	w.callback = callback
	ctx, cancel := context.WithCancel(context.Background())
	w.cancel = cancel

	go func() {
		defer w.reader.Close()
		for {
			msg, err := w.reader.ReadMessage(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				continue
			}
			w.mu.Lock()
			cb := w.callback
			w.mu.Unlock()
			if cb != nil {
				cb(string(msg.Value))
			}
		}
	}()

	return nil
}

// Update publishes a reload event to the topic.
func (w *Watcher) Update() error {
	w.mu.Lock()
	if w.closed {
		w.mu.Unlock()
		return nil
	}
	w.mu.Unlock()

	data, err := json.Marshal(map[string]string{"type": "reload"})
	if err != nil {
		return err
	}
	return w.writer.WriteMessages(context.Background(), Message{Key: []byte("policy-reload"), Value: data})
}

func (w *Watcher) Close() {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.closed {
		return
	}
	w.closed = true
	if w.cancel != nil {
		w.cancel()
	}
	w.writer.Close()
}
