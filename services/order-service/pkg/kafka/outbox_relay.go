package kafka

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
)

type OutboxRelay struct {
	writer       *kafka.Writer
	outbox       OutboxRepo
	pollInterval time.Duration
}

type OutboxRepo interface {
	ListUnprocessed(ctx context.Context, limit int) ([]OutboxEventRow, error)
	MarkProcessed(ctx context.Context, ids []int64) error
}

type OutboxEventRow struct {
	ID      int64
	Topic   string
	Key     string
	Payload []byte
}

func NewOutboxRelay(brokers []string, topic string, outbox OutboxRepo) *OutboxRelay {
	return &OutboxRelay{writer: NewWriter(brokers, topic), outbox: outbox, pollInterval: 5 * time.Second}
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

func DecodeOutboxEvent(data []byte) (OutboxEvent, error) {
	var event OutboxEvent
	err := json.Unmarshal(data, &event)
	return event, err
}
