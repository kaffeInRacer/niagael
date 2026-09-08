package IUseCase

import (
	"context"
	"kaffein/product-service/internal/domain"
	"kaffein/product-service/internal/dto"
)

type CategoryUseCase interface {
	List(ctx context.Context, params dto.ListCategoryParams) ([]domain.Category, int64, error)
	Create(ctx context.Context, args dto.CreateCategoryDto) error
	Update(ctx context.Context, id string, args dto.UpdateCategoryDto) error
	Delete(ctx context.Context, id string) error
	ReadById(ctx context.Context, id string) (*domain.Category, error)
	ReadBySlug(ctx context.Context, slug string) (*domain.Category, error)
}
