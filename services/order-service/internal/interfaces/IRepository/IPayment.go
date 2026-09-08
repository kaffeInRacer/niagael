package IRepository

import (
	"context"
	"kaffein/order-service/internal/domain"
)

type PaymentRepository interface {
	Create(ctx context.Context, args domain.Payment) error
	ReadByOrderId(ctx context.Context, orderId string) (*domain.Payment, error)
	ReadById(ctx context.Context, id string) (*domain.Payment, error)
	UpdateStatus(ctx context.Context, id string, status string) error
}
