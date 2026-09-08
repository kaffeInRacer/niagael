package IRepository

import (
	"context"
	"kaffein/product-service/internal/domain"
	"kaffein/product-service/internal/dto"
)

type ProductRepository interface {
	List(ctx context.Context, params dto.ListProductParams) ([]domain.Product, error)
	ListCount(ctx context.Context, params dto.ListProductParams) (int64, error)
	Create(ctx context.Context, args dto.CreateProductDto) error
	Update(ctx context.Context, args dto.UpdateProductDto) error
	Delete(ctx context.Context, id string) error
	ReadById(ctx context.Context, id string) (*domain.Product, error)
	ReadBySlug(ctx context.Context, slug string) (*domain.Product, error)
	BatchById(ctx context.Context, ids []string) ([]domain.Product, error)
	ReserveStock(ctx context.Context, productId string, variantId string, quantity int32) error
	ReleaseStock(ctx context.Context, productId string, variantId string, quantity int32) error
	ConfirmStock(ctx context.Context, productId string, variantId string, quantity int32) error
}

type ProductImageRepository interface {
	Create(ctx context.Context, args dto.CreateProductImageDto) error
	Delete(ctx context.Context, id string) error
	ListByProductId(ctx context.Context, productId string) ([]domain.ProductImage, error)
	ListByProductIds(ctx context.Context, productIds []string) ([]domain.ProductImage, error)
}
