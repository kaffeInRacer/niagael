package kafka

import (
	"context"
	"encoding/json"
	"time"

	"kaffein/product-service/internal/domain"
)

type FlashSaleEvent struct {
	Type            string     `json:"type"`
	ID              string     `json:"id"`
	Name            string     `json:"name"`
	ProductID       string     `json:"product_id"`
	VariantID       *string    `json:"variant_id"`
	DiscountPercent int        `json:"discount_percent"`
	StartTime       time.Time  `json:"start_time"`
	EndTime         time.Time  `json:"end_time"`
	IsActive        bool       `json:"is_active"`
	DeletedAt       *time.Time `json:"deleted_at"`
}

type FlashSaleProjectionRepo interface {
	Upsert(ctx context.Context, p domain.FlashSaleProjection) error
	Delete(ctx context.Context, productID, variantID string) error
}

type FlashSaleConsumer struct {
	repo FlashSaleProjectionRepo
}

func NewFlashSaleConsumer(repo FlashSaleProjectionRepo) *FlashSaleConsumer {
	return &FlashSaleConsumer{repo: repo}
}

// Handle processes one flash-sale-events message. It is wired to event.Consume.
func (c *FlashSaleConsumer) Handle(ctx context.Context, data []byte) error {
	var event FlashSaleEvent
	if err := json.Unmarshal(data, &event); err != nil {
		return nil
	}

	variantID := ""
	if event.VariantID != nil {
		variantID = *event.VariantID
	}

	if event.Type == "deleted" || event.DeletedAt != nil {
		return c.repo.Delete(ctx, event.ProductID, variantID)
	}

	return c.repo.Upsert(ctx, domain.FlashSaleProjection{
		ProductID:       event.ProductID,
		VariantID:       variantID,
		Name:            event.Name,
		DiscountPercent: event.DiscountPercent,
		StartTime:       event.StartTime,
		EndTime:         event.EndTime,
		IsActive:        event.IsActive,
	})
}
