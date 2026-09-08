package IRepository

import (
	"context"
	"kaffein/order-service/internal/domain"
)

type OrderItemRepository interface {
	Create(ctx context.Context, args domain.OrderItem) error
	ReadByOrderId(ctx context.Context, orderId string) ([]domain.OrderItem, error)
}
