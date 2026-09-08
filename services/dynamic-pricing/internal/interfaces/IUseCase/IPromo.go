package IUseCase

import (
	"context"
	"kaffein/dynamic-pricing-service/internal/domain"
	"kaffein/dynamic-pricing-service/internal/dto"
)

type PromoUseCase interface {
	List(ctx context.Context, params dto.ListPromoParams) ([]domain.Promo, int64, error)
	Create(ctx context.Context, args dto.CreatePromoDto) error
	Update(ctx context.Context, id string, args dto.UpdatePromoDto) error
	Delete(ctx context.Context, id string) error
	ReadById(ctx context.Context, id string) (*domain.Promo, error)
	ReadByCode(ctx context.Context, code string) (*domain.Promo, error)
	ApplyPromo(ctx context.Context, code string, userId string, totalAmount int64) (int64, error)
	GetUsage(ctx context.Context, promoId string, userId string) (*domain.PromoUsage, error)
}
