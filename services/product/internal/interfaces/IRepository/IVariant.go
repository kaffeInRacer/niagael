package IRepository

import (
	"context"
	"kaffein/product-service/internal/domain"
	"kaffein/product-service/internal/dto"
)

type VariantRepository interface {
	ListByProductId(ctx context.Context, productId string) ([]domain.ProductVariant, error)
	ListByProductIds(ctx context.Context, productIds []string) ([]domain.ProductVariant, error)
	Create(ctx context.Context, args dto.CreateVariantDto) error
	Update(ctx context.Context, args dto.UpdateVariantDto) error
	Delete(ctx context.Context, id string) error
	ReadById(ctx context.Context, id string) (*domain.ProductVariant, error)
	ReserveStock(ctx context.Context, variantId string, quantity int32) error
	ReleaseStock(ctx context.Context, variantId string, quantity int32) error
	ConfirmStock(ctx context.Context, variantId string, quantity int32) error
}
