package usecase

import (
	"context"
	"kaffein/order-service/internal/domain"
	"kaffein/order-service/internal/dto"
	"kaffein/order-service/internal/interfaces/IRepository"
	"kaffein/order-service/internal/interfaces/IUseCase"
	"time"

	"github.com/google/uuid"
)

type addressUseCase struct {
	repo IRepository.AddressRepository
}

func NewAddressUseCase(repo IRepository.AddressRepository) IUseCase.AddressUseCase {
	return &addressUseCase{repo: repo}
}

func (uc *addressUseCase) List(ctx context.Context, params dto.ListAddressParams) ([]domain.Address, error) {
	return uc.repo.List(ctx, params)
}

func (uc *addressUseCase) Create(ctx context.Context, args dto.CreateAddressDto) (*domain.Address, error) {
	arg := dto.CreateAddressDto{
		Id:         uuid.New().String(),
		UserId:     args.UserId,
		Street:     args.Street,
		City:       args.City,
		Province:   args.Province,
		PostalCode: args.PostalCode,
		Country:    args.Country,
		IsDefault:  args.IsDefault,
	}

	if err := uc.repo.Create(ctx, arg); err != nil {
		return nil, err
	}

	return &domain.Address{
		Id:         arg.Id,
		UserId:     arg.UserId,
		Street:     arg.Street,
		City:       arg.City,
		Province:   arg.Province,
		PostalCode: arg.PostalCode,
		Country:    arg.Country,
		IsDefault:  arg.IsDefault,
		CreatedAt:  time.Now().Format(time.RFC3339),
	}, nil
}

func (uc *addressUseCase) Update(ctx context.Context, id string, args dto.UpdateAddressDto) error {
	arg := dto.UpdateAddressDto{
		UserId:     args.UserId,
		Street:     args.Street,
		City:       args.City,
		Province:   args.Province,
		PostalCode: args.PostalCode,
		Country:    args.Country,
		IsDefault:  args.IsDefault,
	}

	return uc.repo.Update(ctx, id, arg)
}

func (uc *addressUseCase) Delete(ctx context.Context, id string) error {
	return uc.repo.Delete(ctx, id)
}

func (uc *addressUseCase) ReadById(ctx context.Context, id string) (*domain.Address, error) {
	return uc.repo.ReadById(ctx, id)
}
