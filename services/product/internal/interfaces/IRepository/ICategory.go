package IRepository

import (
	"context"
	"kaffein/product-service/internal/domain"
	"kaffein/product-service/internal/dto"
)

type CategoryRepository interface {
	List(ctx context.Context, params dto.ListCategoryParams) ([]domain.Category, error)
	ListCount(ctx context.Context, params dto.ListCategoryParams) (int64, error)
	ListWithProductCount(ctx context.Context, params dto.ListCategoryParams) ([]domain.Category, error)
	Create(ctx context.Context, args dto.CreateCategoryDto) error
	Update(ctx context.Context, args dto.UpdateCategoryDto) error
	Delete(ctx context.Context, id string) error
	ReadById(ctx context.Context, id string) (*domain.Category, error)
	ReadBySlug(ctx context.Context, slug string) (*domain.Category, error)
}
