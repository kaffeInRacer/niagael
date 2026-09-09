package kafka

import (
	"context"
	"encoding/json"
	"log"

	"github.com/segmentio/kafka-go"
)

type UserEvent struct {
	Type     string `json:"type"`
	ID       string `json:"id"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	IsActive *bool  `json:"is_active"`
}

type BuyersRepo interface {
	Upsert(ctx context.Context, id, email string) error
	Delete(ctx context.Context, id string) error
}

type CasbinPolicyConsumer struct {
	reader *kafka.Reader
	buyers BuyersRepo
}

func NewCasbinPolicyConsumer(brokers []string, topic, groupID string, buyers BuyersRepo) *CasbinPolicyConsumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:               brokers,
		Topic:                 topic,
		GroupID:               groupID,
		MinBytes:              1,
		MaxBytes:              10e6,
		WatchPartitionChanges: true,
	})
	return &CasbinPolicyConsumer{reader: reader, buyers: buyers}
}

func (c *CasbinPolicyConsumer) Run(ctx context.Context) {
	go func() {
		for {
			msg, err := c.reader.ReadMessage(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				log.Printf("kafka user-events read error: %v", err)
				continue
			}
			if err := c.handle(ctx, msg.Value); err != nil {
				log.Printf("kafka user-events handle error: %v", err)
			}
		}
	}()
}

func (c *CasbinPolicyConsumer) handle(ctx context.Context, data []byte) error {
	var event UserEvent
	if err := json.Unmarshal(data, &event); err != nil {
		return nil
	}

	if event.Type == "deleted" {
		return c.buyers.Delete(ctx, event.ID)
	}
	if event.Email == "" {
		return nil
	}
	return c.buyers.Upsert(ctx, event.ID, event.Email)
}

func (c *CasbinPolicyConsumer) Close() error {
	return c.reader.Close()
}
