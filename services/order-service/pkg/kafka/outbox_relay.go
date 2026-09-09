package kafka

import (
	"context"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
)

type OutboxRelay struct {
	writer       *kafka.Writer
	outbox       OutboxRepo
	pollInterval time.Duration
}

func NewOutboxRelay(brokers []string, topic string, outbox OutboxRepo) *OutboxRelay {
	writer := &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Topic:        topic,
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: kafka.RequireAll,
	}
	return &OutboxRelay{writer: writer, outbox: outbox, pollInterval: 5 * time.Second}
}

func (r *OutboxRelay) Run(ctx context.Context) {
	ticker := time.NewTicker(r.pollInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			r.relayBatch(ctx)
		}
	}
}

func (r *OutboxRelay) relayBatch(ctx context.Context) {
	events, err := r.outbox.ListUnprocessed(ctx, 100)
	if err != nil {
		log.Printf("outbox relay query error: %v", err)
		return
	}
	if len(events) == 0 {
		return
	}

	messages := make([]kafka.Message, 0, len(events))
	ids := make([]int64, 0, len(events))
	for _, e := range events {
		messages = append(messages, kafka.Message{Key: []byte(e.Key), Value: e.Payload})
		ids = append(ids, e.ID)
	}

	if err := r.writer.WriteMessages(ctx, messages...); err != nil {
		log.Printf("outbox relay publish error (will retry): %v", err)
		return
	}
	if err := r.outbox.MarkProcessed(ctx, ids); err != nil {
		log.Printf("outbox relay delete error: %v", err)
	}
}

func (r *OutboxRelay) Close() error {
	return r.writer.Close()
}
