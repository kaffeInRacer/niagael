package IRepository

import (
	"context"
	"kaffein/order-service/internal/domain"
	"kaffein/order-service/internal/dto"
)

type AddressRepository interface {
	List(ctx context.Context, params dto.ListAddressParams) ([]domain.Address, error)
	Create(ctx context.Context, args dto.CreateAddressDto) error
	Update(ctx context.Context, id string, args dto.UpdateAddressDto) error
	Delete(ctx context.Context, id string) error
	ReadById(ctx context.Context, id string) (*domain.Address, error)
}
