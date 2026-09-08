package IRepository

import (
	"context"
	"kaffein/dynamic-pricing-service/internal/domain"
)

type FlashSaleUsageRepository interface {
	GetUsage(ctx context.Context, flashSaleId string, userId string) (*domain.FlashSaleUsage, error)
	IncrementUsage(ctx context.Context, flashSaleId string, userId string, quantity int) error
}
