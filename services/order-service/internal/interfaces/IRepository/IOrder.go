package IRepository

import (
	"context"
	"kaffein/order-service/internal/domain"
	"kaffein/order-service/internal/dto"
)

type OrderRepository interface {
	Create(ctx context.Context, args domain.Order) error
	CreateWithItemsWithTx(ctx context.Context, order domain.Order, items []domain.OrderItem) error
	Update(ctx context.Context, id string, args domain.Order) error
	UpdateStatus(ctx context.Context, id string, fromStatuses []string, status string) error
	ReadById(ctx context.Context, id string) (*domain.Order, error)
	ReadItemsByOrderId(ctx context.Context, orderId string) ([]domain.OrderItem, error)
	ReadByUserId(ctx context.Context, userId string, before string, pageSize int32) ([]domain.Order, error)
	List(ctx context.Context, params dto.ListOrderParams) ([]domain.Order, error)
	ListCount(ctx context.Context, params dto.ListOrderParams) (int64, error)
}
