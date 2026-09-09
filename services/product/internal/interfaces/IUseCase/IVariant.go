package IUseCase

import (
	"context"
	"kaffein/product-service/internal/domain"
	"kaffein/product-service/internal/dto"
)

type VariantUseCase interface {
	ListByProductId(ctx context.Context, productId string) ([]domain.ProductVariant, error)
	ListActiveByProductId(ctx context.Context, productId string) ([]domain.ProductVariant, error)
	Create(ctx context.Context, productId string, args dto.CreateVariantDto) error
	Update(ctx context.Context, id string, args dto.UpdateVariantDto) error
	Delete(ctx context.Context, id string) error
	ReadById(ctx context.Context, id string) (*domain.ProductVariant, error)
}
