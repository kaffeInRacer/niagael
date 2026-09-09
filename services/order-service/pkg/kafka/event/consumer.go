package event

import (
	"context"
	"log"

	"github.com/segmentio/kafka-go"
)

// Consume runs a blocking loop that reads messages from the topic and calls
// handle for each one. It is meant to be started in its own goroutine.
func Consume(ctx context.Context, brokers []string, topic, groupID string, handle func(ctx context.Context, data []byte) error) {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:               brokers,
		Topic:                 topic,
		GroupID:               groupID,
		MinBytes:              1,
		MaxBytes:              10e6,
		WatchPartitionChanges: true,
	})
	defer r.Close()
	for {
		msg, err := r.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			log.Printf("kafka %s read error: %v", topic, err)
			continue
		}
		if err := handle(ctx, msg.Value); err != nil {
			log.Printf("kafka %s handle error: %v", topic, err)
		}
	}
}
