package IUseCase

import (
	"context"
	"kaffein/product-service/internal/domain"
	"kaffein/product-service/internal/dto"
)

type ProductUseCase interface {
	List(ctx context.Context, params dto.ListProductParams) ([]domain.Product, int64, error)
	Create(ctx context.Context, args dto.CreateProductDto) error
	Update(ctx context.Context, id string, args dto.UpdateProductDto) error
	Delete(ctx context.Context, id string) error
	ReadById(ctx context.Context, id string) (*domain.Product, error)
	ReadBySlug(ctx context.Context, slug string) (*domain.Product, error)
	BatchById(ctx context.Context, ids []string) ([]domain.Product, error)
	ReserveStock(ctx context.Context, productId string, variantId string, quantity int32) error
	ReleaseStock(ctx context.Context, productId string, variantId string, quantity int32) error
	ConfirmStock(ctx context.Context, productId string, variantId string, quantity int32) error
}

type ProductImageUseCase interface {
	Create(ctx context.Context, productId string, fileName string, sortOrder int) error
	Delete(ctx context.Context, id string) error
	ListByProductId(ctx context.Context, productId string) ([]domain.ProductImage, error)
}
