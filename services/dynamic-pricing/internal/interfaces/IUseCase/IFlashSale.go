package IUseCase

import (
	"context"
	"kaffein/dynamic-pricing-service/internal/domain"
	"kaffein/dynamic-pricing-service/internal/dto"
)

type FlashSaleUseCase interface {
	List(ctx context.Context, params dto.ListFlashSaleParams) ([]domain.FlashSale, int64, error)
	Create(ctx context.Context, args dto.CreateFlashSaleDto) error
	CreateBulk(ctx context.Context, args dto.CreateFlashSaleBulkDto) error
	Update(ctx context.Context, id string, args dto.UpdateFlashSaleDto) error
	Delete(ctx context.Context, id string) error
	ReadById(ctx context.Context, id string) (*domain.FlashSale, error)
	ReadByProductId(ctx context.Context, productId string) (*domain.FlashSale, error)
	ReadByVariantId(ctx context.Context, productId string, variantId string) (*domain.FlashSale, error)
	ReadByProductIds(ctx context.Context, productIds []string) ([]domain.FlashSale, error)
	ReadByVariantIds(ctx context.Context, items []dto.VariantIdPair) ([]domain.FlashSale, error)
	DecrementStock(ctx context.Context, id string, quantity int) error
	IncrementUsage(ctx context.Context, flashSaleId string, userId string, quantity int) error
}
