package kafka

import (
	"context"
	"encoding/json"
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

type BuyersConsumer struct {
	buyers BuyersRepo
}

func NewBuyersConsumer(buyers BuyersRepo) *BuyersConsumer {
	return &BuyersConsumer{buyers: buyers}
}

// Handle processes one user-events message. It is wired to event.Consume.
func (c *BuyersConsumer) Handle(ctx context.Context, data []byte) error {
	var e UserEvent
	if err := json.Unmarshal(data, &e); err != nil {
		return nil
	}
	if e.Type == "deleted" {
		return c.buyers.Delete(ctx, e.ID)
	}
	if e.Email == "" {
		return nil
	}
	return c.buyers.Upsert(ctx, e.ID, e.Email)
}
