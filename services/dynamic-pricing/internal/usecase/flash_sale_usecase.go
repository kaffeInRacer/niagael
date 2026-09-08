package usecase

import (
	"context"
	"kaffein/dynamic-pricing-service/internal/domain"
	"kaffein/dynamic-pricing-service/internal/dto"
	"kaffein/dynamic-pricing-service/internal/interfaces/IRepository"

	"github.com/google/uuid"
)

type flashSaleUseCase struct {
	repo      IRepository.FlashSaleRepository
	usageRepo IRepository.FlashSaleUsageRepository
}

func NewFlashSaleUseCase(repo IRepository.FlashSaleRepository, usageRepo IRepository.FlashSaleUsageRepository) *flashSaleUseCase {
	return &flashSaleUseCase{repo: repo, usageRepo: usageRepo}
}

func (uc *flashSaleUseCase) List(ctx context.Context, params dto.ListFlashSaleParams) ([]domain.FlashSale, int64, error) {
	flashSales, err := uc.repo.List(ctx, params)
	if err != nil {
		return nil, 0, err
	}

	count, err := uc.repo.ListCount(ctx, params)
	if err != nil {
		return nil, 0, err
	}

	return flashSales, count, nil
}

func (uc *flashSaleUseCase) Create(ctx context.Context, args dto.CreateFlashSaleDto) error {
	arg := dto.CreateFlashSaleDto{
		Id:              uuid.New().String(),
		Name:            args.Name,
		ProductId:       args.ProductId,
		VariantId:       args.VariantId,
		DiscountPercent: args.DiscountPercent,
		Stock:           args.Stock,
		MaxPerUser:      args.MaxPerUser,
		StartTime:       args.StartTime,
		EndTime:         args.EndTime,
		IsActive:        args.IsActive,
	}

	return uc.repo.Create(ctx, arg)
}

func (uc *flashSaleUseCase) CreateBulk(ctx context.Context, args dto.CreateFlashSaleBulkDto) error {
	items := make([]dto.CreateFlashSaleItemDto, len(args.Items))
	for i, item := range args.Items {
		items[i] = dto.CreateFlashSaleItemDto{
			Id:              uuid.New().String(),
			ProductId:       item.ProductId,
			VariantId:       item.VariantId,
			DiscountPercent: item.DiscountPercent,
			Stock:           item.Stock,
			MaxPerUser:      item.MaxPerUser,
		}
	}

	bulk := dto.CreateFlashSaleBulkDto{
		Name:      args.Name,
		StartTime: args.StartTime,
		EndTime:   args.EndTime,
		IsActive:  args.IsActive,
		Items:     items,
	}

	return uc.repo.CreateBulk(ctx, bulk)
}

func (uc *flashSaleUseCase) Update(ctx context.Context, id string, args dto.UpdateFlashSaleDto) error {
	arg := dto.UpdateFlashSaleDto{
		Id:              id,
		Name:            args.Name,
		ProductId:       args.ProductId,
		VariantId:       args.VariantId,
		DiscountPercent: args.DiscountPercent,
		Stock:           args.Stock,
		MaxPerUser:      args.MaxPerUser,
		StartTime:       args.StartTime,
		EndTime:         args.EndTime,
		IsActive:        args.IsActive,
	}

	return uc.repo.Update(ctx, arg)
}

func (uc *flashSaleUseCase) Delete(ctx context.Context, id string) error {
	return uc.repo.Delete(ctx, id)
}

func (uc *flashSaleUseCase) ReadById(ctx context.Context, id string) (*domain.FlashSale, error) {
	return uc.repo.ReadById(ctx, id)
}

func (uc *flashSaleUseCase) ReadByProductId(ctx context.Context, productId string) (*domain.FlashSale, error) {
	return uc.repo.ReadByProductId(ctx, productId)
}

func (uc *flashSaleUseCase) ReadByVariantId(ctx context.Context, productId string, variantId string) (*domain.FlashSale, error) {
	return uc.repo.ReadByVariantId(ctx, productId, variantId)
}

func (uc *flashSaleUseCase) ReadByProductIds(ctx context.Context, productIds []string) ([]domain.FlashSale, error) {
	return uc.repo.ReadByProductIds(ctx, productIds)
}

func (uc *flashSaleUseCase) ReadByVariantIds(ctx context.Context, items []dto.VariantIdPair) ([]domain.FlashSale, error) {
	return uc.repo.ReadByVariantIds(ctx, items)
}

func (uc *flashSaleUseCase) DecrementStock(ctx context.Context, id string, quantity int) error {
	return uc.repo.DecrementStock(ctx, id, quantity)
}

func (uc *flashSaleUseCase) IncrementUsage(ctx context.Context, flashSaleId string, userId string, quantity int) error {
	return uc.usageRepo.IncrementUsage(ctx, flashSaleId, userId, quantity)
}
