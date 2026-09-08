package IRepository

import (
	"context"
	"kaffein/dynamic-pricing-service/internal/domain"
)

type PromoUsageRepository interface {
	GetUsage(ctx context.Context, promoId string, userId string) (*domain.PromoUsage, error)
	IncrementUsage(ctx context.Context, promoId string, userId string, quantity int) error
}
