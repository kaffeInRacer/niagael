package usecase

import (
	"context"
	"errors"
	"kaffein/dynamic-pricing-service/internal/domain"
	"kaffein/dynamic-pricing-service/internal/dto"
	"kaffein/dynamic-pricing-service/internal/interfaces/IRepository"
	"kaffein/dynamic-pricing-service/utils/constants"
	"time"

	"github.com/google/uuid"
)

type promoUseCase struct {
	repo      IRepository.PromoRepository
	usageRepo IRepository.PromoUsageRepository
}

func NewPromoUseCase(repo IRepository.PromoRepository, usageRepo IRepository.PromoUsageRepository) *promoUseCase {
	return &promoUseCase{repo: repo, usageRepo: usageRepo}
}

func (uc *promoUseCase) List(ctx context.Context, params dto.ListPromoParams) ([]domain.Promo, int64, error) {
	promos, err := uc.repo.List(ctx, params)
	if err != nil {
		return nil, 0, err
	}

	count, err := uc.repo.ListCount(ctx, params)
	if err != nil {
		return nil, 0, err
	}

	return promos, count, nil
}

func (uc *promoUseCase) Create(ctx context.Context, args dto.CreatePromoDto) error {
	arg := dto.CreatePromoDto{
		Id:                  uuid.New().String(),
		Name:                args.Name,
		Code:                args.Code,
		Description:         args.Description,
		DiscountType:        args.DiscountType,
		DiscountValue:       args.DiscountValue,
		MinPurchase:         args.MinPurchase,
		MaxDiscount:         args.MaxDiscount,
		Quantity:            args.Quantity,
		MaxUsagePerUser:     args.MaxUsagePerUser,
		CanCombineFlashSale: args.CanCombineFlashSale,
		StartDate:           args.StartDate,
		EndDate:             args.EndDate,
		IsActive:            args.IsActive,
	}

	return uc.repo.Create(ctx, arg)
}

func (uc *promoUseCase) Update(ctx context.Context, id string, args dto.UpdatePromoDto) error {
	arg := dto.UpdatePromoDto{
		Id:                  id,
		Name:                args.Name,
		Code:                args.Code,
		Description:         args.Description,
		DiscountType:        args.DiscountType,
		DiscountValue:       args.DiscountValue,
		MinPurchase:         args.MinPurchase,
		MaxDiscount:         args.MaxDiscount,
		Quantity:            args.Quantity,
		MaxUsagePerUser:     args.MaxUsagePerUser,
		CanCombineFlashSale: args.CanCombineFlashSale,
		StartDate:           args.StartDate,
		EndDate:             args.EndDate,
		IsActive:            args.IsActive,
	}

	return uc.repo.Update(ctx, arg)
}

func (uc *promoUseCase) Delete(ctx context.Context, id string) error {
	return uc.repo.Delete(ctx, id)
}

func (uc *promoUseCase) ReadById(ctx context.Context, id string) (*domain.Promo, error) {
	return uc.repo.ReadById(ctx, id)
}

func (uc *promoUseCase) ReadByCode(ctx context.Context, code string) (*domain.Promo, error) {
	return uc.repo.ReadByCode(ctx, code)
}

func (uc *promoUseCase) ApplyPromo(ctx context.Context, code string, userId string, totalAmount int64) (int64, error) {
	promo, err := uc.repo.ReadByCode(ctx, code)
	if err != nil {
		return 0, err
	}

	if promo == nil {
		return 0, errors.New(constants.ErrPromoNotFound)
	}

	if !promo.IsActive {
		return 0, errors.New(constants.ErrPromoNotActive)
	}

	now := time.Now()
	if now.Before(promo.StartDate) || now.After(promo.EndDate) {
		return 0, errors.New(constants.ErrPromoExpired)
	}

	if promo.Quantity > 0 && promo.UsedCount >= promo.Quantity {
		return 0, errors.New(constants.ErrPromoLimitReached)
	}

	if promo.MaxUsagePerUser > 0 {
		usage, err := uc.usageRepo.GetUsage(ctx, promo.Id, userId)
		if err == nil && usage != nil && usage.Quantity >= promo.MaxUsagePerUser {
			return 0, errors.New(constants.ErrPromoMaxUsagePerUser)
		}
	}

	if totalAmount < promo.MinPurchase {
		return 0, errors.New(constants.ErrPromoMinPurchase)
	}

	var discount int64
	switch promo.DiscountType {
	case "percentage":
		discount = totalAmount * promo.DiscountValue / 100
		if promo.MaxDiscount > 0 && discount > promo.MaxDiscount {
			discount = promo.MaxDiscount
		}
	case "fixed":
		discount = promo.DiscountValue
		if discount > totalAmount {
			discount = totalAmount
		}
	}

	// Applying a code in the cart is a quote and must not consume promo quota.
	return discount, nil
}

func (uc *promoUseCase) GetUsage(ctx context.Context, promoId string, userId string) (*domain.PromoUsage, error) {
	usage, err := uc.usageRepo.GetUsage(ctx, promoId, userId)
	if err != nil {
		// If no usage found, return nil (not an error)
		return nil, nil
	}
	return usage, nil
}
