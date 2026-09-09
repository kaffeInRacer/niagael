package usecase

import (
	"context"
	"errors"
	"kaffein/dynamic-pricing-service/internal/domain"
	"kaffein/dynamic-pricing-service/internal/dto"
	"kaffein/dynamic-pricing-service/internal/interfaces/IRepository"
	"kaffein/dynamic-pricing-service/utils/constants"
	"kaffein/dynamic-pricing-service/utils/events"

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

	if err := uc.repo.Create(ctx, arg); err != nil {
		return err
	}
	publishFlashSaleChange("updated", map[string]any{
		"id": arg.Id, "name": arg.Name, "product_id": arg.ProductId,
		"variant_id": arg.VariantId, "discount_percent": arg.DiscountPercent,
		"start_time": arg.StartTime, "end_time": arg.EndTime, "is_active": arg.IsActive,
	})
	return nil
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

	if err := uc.repo.CreateBulk(ctx, bulk); err != nil {
		return err
	}
	for _, item := range bulk.Items {
		publishFlashSaleChange("updated", map[string]any{
			"id": item.Id, "name": bulk.Name, "product_id": item.ProductId,
			"variant_id": item.VariantId, "discount_percent": item.DiscountPercent,
			"start_time": bulk.StartTime, "end_time": bulk.EndTime, "is_active": bulk.IsActive,
		})
	}
	return nil
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

	if err := uc.repo.Update(ctx, arg); err != nil {
		return err
	}
	publishFlashSaleChange("updated", map[string]any{
		"id": arg.Id, "name": arg.Name, "product_id": arg.ProductId,
		"variant_id": arg.VariantId, "discount_percent": arg.DiscountPercent,
		"start_time": arg.StartTime, "end_time": arg.EndTime, "is_active": arg.IsActive,
	})
	return nil
}

func (uc *flashSaleUseCase) Delete(ctx context.Context, id string) error {
	fs, err := uc.repo.ReadById(ctx, id)
	if err != nil {
		return err
	}
	if err := uc.repo.Delete(ctx, id); err != nil {
		return err
	}
	if fs != nil {
		publishFlashSaleChange("deleted", map[string]any{
			"id": id, "product_id": fs.ProductId, "variant_id": fs.VariantId,
		})
	}
	return nil
}

func publishFlashSaleChange(eventType string, payload map[string]any) {
	payload["type"] = eventType
	events.Publish(constants.FlashSaleTopic, "flash-sale-changed", payload)
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
	if quantity <= 0 {
		return errors.New(constants.ErrFlashSaleQuantityInvalid)
	}
	return uc.repo.DecrementStock(ctx, id, quantity)
}

func (uc *flashSaleUseCase) IncrementUsage(ctx context.Context, flashSaleId string, userId string, quantity int) error {
	if quantity <= 0 {
		return errors.New(constants.ErrFlashSaleQuantityInvalid)
	}
	return uc.usageRepo.IncrementUsage(ctx, flashSaleId, userId, quantity)
}
