package event

import (
	"context"
	"encoding/json"

	"github.com/segmentio/kafka-go"
)

// Publish marshals the payload and writes it to the topic. It is a no-op when
// brokers is empty so callers do not need feature checks.
func Publish(brokers []string, topic, key string, payload any) {
	if len(brokers) == 0 {
		return
	}
	w := &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Topic:        topic,
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: kafka.RequireAll,
	}
	defer w.Close()
	data, err := json.Marshal(payload)
	if err != nil {
		return
	}
	_ = w.WriteMessages(context.Background(), kafka.Message{Key: []byte(key), Value: data})
}
