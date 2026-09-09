package usecase

import (
	"context"
	"kaffein/product-service/internal/domain"
	"kaffein/product-service/internal/dto"
	"kaffein/product-service/internal/interfaces/IRepository"

	"github.com/google/uuid"
)

type productImageUseCase struct {
	repo IRepository.ProductImageRepository
}

func NewProductImageUseCase(repo IRepository.ProductImageRepository) *productImageUseCase {
	return &productImageUseCase{repo: repo}
}

func (uc *productImageUseCase) Create(ctx context.Context, productId string, fileName string, sortOrder int) error {
	arg := dto.CreateProductImageDto{
		Id:        uuid.New().String(),
		ProductId: productId,
		FileName:  fileName,
		SortOrder: sortOrder,
	}

	return uc.repo.Create(ctx, arg)
}

func (uc *productImageUseCase) Delete(ctx context.Context, productId string, id string) error {
	return uc.repo.Delete(ctx, productId, id)
}

func (uc *productImageUseCase) ListByProductId(ctx context.Context, productId string) ([]domain.ProductImage, error) {
	return uc.repo.ListByProductId(ctx, productId)
}
