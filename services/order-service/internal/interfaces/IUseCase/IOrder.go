package IUseCase

import (
	"context"
	"kaffein/order-service/internal/domain"
	"kaffein/order-service/internal/dto"
)

type OrderUseCase interface {
	Create(ctx context.Context, args dto.CreateOrderDto) (*domain.Order, error)
	List(ctx context.Context, params dto.ListOrderParams) ([]domain.Order, int64, error)
	ReadById(ctx context.Context, id string) (*domain.Order, error)
	ReadByIdWithItems(ctx context.Context, id string) (*domain.Order, []domain.OrderItem, error)
	ReadByUserId(ctx context.Context, userId string, before string, pageSize int32) ([]domain.Order, error)
	UpdateStatus(ctx context.Context, id string, status string) error
}
