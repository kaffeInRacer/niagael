package IRepository

import (
	"context"
	"kaffein/dynamic-pricing-service/internal/domain"
	"kaffein/dynamic-pricing-service/internal/dto"
)

type PromoRepository interface {
	List(ctx context.Context, params dto.ListPromoParams) ([]domain.Promo, error)
	ListCount(ctx context.Context, params dto.ListPromoParams) (int64, error)
	Create(ctx context.Context, args dto.CreatePromoDto) error
	Update(ctx context.Context, args dto.UpdatePromoDto) error
	Delete(ctx context.Context, id string) error
	ReadById(ctx context.Context, id string) (*domain.Promo, error)
	ReadByCode(ctx context.Context, code string) (*domain.Promo, error)
	IncrementUsage(ctx context.Context, id string) error
}
