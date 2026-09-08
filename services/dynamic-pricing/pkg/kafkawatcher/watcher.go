package kafkawatcher

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/segmentio/kafka-go"
)

type PolicyEvent struct {
	Type     string `json:"type"`
	Service  string `json:"service"`
	Role     string `json:"role"`
	Resource string `json:"resource"`
	Action   string `json:"action"`
}

type Watcher struct {
	brokers  []string
	topic    string
	groupID  string
	callback func(string)
	writer   *kafka.Writer
	cancel   context.CancelFunc
	mu       sync.Mutex
	closed   bool
}

func New(brokers []string, topic, groupID string) *Watcher {
	w := &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Topic:        topic,
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: kafka.RequireAll,
	}
	return &Watcher{
		brokers: brokers,
		topic:   topic,
		groupID: groupID,
		writer:  w,
	}
}

func (w *Watcher) SetUpdateCallback(callback func(string)) error {
	w.callback = callback
	ctx, cancel := context.WithCancel(context.Background())
	w.cancel = cancel

	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:               w.brokers,
		Topic:                 w.topic,
		GroupID:               w.groupID,
		MinBytes:              1,
		MaxBytes:              10e6,
		WatchPartitionChanges: true,
	})

	go func() {
		defer r.Close()
		for {
			msg, err := r.ReadMessage(ctx)
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

func (w *Watcher) Update() error {
	w.mu.Lock()
	if w.closed {
		w.mu.Unlock()
		return nil
	}
	w.mu.Unlock()

	event := PolicyEvent{Type: "reload"}
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}
	return w.writer.WriteMessages(context.Background(), kafka.Message{
		Key:   []byte("policy-reload"),
		Value: data,
	})
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
