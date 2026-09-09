package usecase

import (
	"context"
	"kaffein/product-service/internal/domain"
	"kaffein/product-service/internal/dto"
	"kaffein/product-service/internal/interfaces/IRepository"

	"github.com/google/uuid"
)

type variantUseCase struct {
	repo IRepository.VariantRepository
}

func NewVariantUseCase(repo IRepository.VariantRepository) *variantUseCase {
	return &variantUseCase{repo: repo}
}

func (uc *variantUseCase) ListByProductId(ctx context.Context, productId string) ([]domain.ProductVariant, error) {
	return uc.repo.ListByProductId(ctx, productId)
}

func (uc *variantUseCase) ListActiveByProductId(ctx context.Context, productId string) ([]domain.ProductVariant, error) {
	return uc.repo.ListActiveByProductId(ctx, productId)
}

func (uc *variantUseCase) Create(ctx context.Context, productId string, args dto.CreateVariantDto) error {
	arg := dto.CreateVariantDto{
		Id:         uuid.New().String(),
		ProductId:  productId,
		Name:       args.Name,
		Price:      args.Price,
		Stock:      args.Stock,
		Attributes: args.Attributes,
		IsActive:   args.IsActive,
	}

	return uc.repo.Create(ctx, arg)
}

func (uc *variantUseCase) Update(ctx context.Context, id string, args dto.UpdateVariantDto) error {
	arg := dto.UpdateVariantDto{
		Id:         id,
		Name:       args.Name,
		Price:      args.Price,
		Stock:      args.Stock,
		Attributes: args.Attributes,
		IsActive:   args.IsActive,
	}

	return uc.repo.Update(ctx, arg)
}

func (uc *variantUseCase) Delete(ctx context.Context, id string) error {
	return uc.repo.Delete(ctx, id)
}

func (uc *variantUseCase) ReadById(ctx context.Context, id string) (*domain.ProductVariant, error) {
	return uc.repo.ReadById(ctx, id)
}
