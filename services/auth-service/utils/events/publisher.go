package events

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/segmentio/kafka-go"
)

var (
	mu      sync.Mutex
	brokers []string
	writers = map[string]*kafka.Writer{}
)

func Init(b []string) {
	brokers = b
}

func Publish(topic, key string, payload any) {
	if len(brokers) == 0 {
		return
	}
	w := writerFor(topic)
	data, err := json.Marshal(payload)
	if err != nil {
		return
	}
	_ = w.WriteMessages(context.Background(), kafka.Message{Key: []byte(key), Value: data})
}

func writerFor(topic string) *kafka.Writer {
	mu.Lock()
	defer mu.Unlock()
	if w, ok := writers[topic]; ok {
		return w
	}
	w := &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Topic:        topic,
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: kafka.RequireAll,
	}
	writers[topic] = w
	return w
}
