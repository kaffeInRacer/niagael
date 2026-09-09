package kafka

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
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
	Upsert(ctx context.Context, p FlashSaleProjection) error
	Delete(ctx context.Context, productID, variantID string) error
}

type FlashSaleProjection struct {
	ProductID       string
	VariantID       string
	Name            string
	DiscountPercent int
	StartTime       time.Time
	EndTime         time.Time
	IsActive        bool
}

type FlashSaleConsumer struct {
	reader *kafka.Reader
	repo   FlashSaleProjectionRepo
}

func NewFlashSaleConsumer(brokers []string, topic, groupID string, repo FlashSaleProjectionRepo) *FlashSaleConsumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:               brokers,
		Topic:                 topic,
		GroupID:               groupID,
		MinBytes:              1,
		MaxBytes:              10e6,
		WatchPartitionChanges: true,
	})
	return &FlashSaleConsumer{reader: reader, repo: repo}
}

func (c *FlashSaleConsumer) Run(ctx context.Context) {
	go func() {
		for {
			msg, err := c.reader.ReadMessage(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				log.Printf("kafka flash-sale-events read error: %v", err)
				continue
			}
			if err := c.handle(ctx, msg.Value); err != nil {
				log.Printf("kafka flash-sale-events handle error: %v", err)
			}
		}
	}()
}

func (c *FlashSaleConsumer) handle(ctx context.Context, data []byte) error {
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

	return c.repo.Upsert(ctx, FlashSaleProjection{
		ProductID:       event.ProductID,
		VariantID:       variantID,
		Name:            event.Name,
		DiscountPercent: event.DiscountPercent,
		StartTime:       event.StartTime,
		EndTime:         event.EndTime,
		IsActive:        event.IsActive,
	})
}

func (c *FlashSaleConsumer) Close() error {
	return c.reader.Close()
}
