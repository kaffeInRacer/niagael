package IRepository

import (
	"context"

	"kaffein/order-service/internal/domain"
	"kaffein/order-service/pkg/postgresql"
)

type PaymentRepository interface {
	GetOrCreate(ctx context.Context, args domain.Payment) (*domain.Payment, error)
	InitiateWithTx(ctx context.Context, orderId string, createTransaction func(*domain.Payment) (string, string, error)) (string, error)
	ReadByOrderId(ctx context.Context, orderId string) (*domain.Payment, error)
	ReadById(ctx context.Context, id string) (*domain.Payment, error)
	ProcessCallbackWithTx(ctx context.Context, orderId string, amount int64, paymentStatus string, stockAction func(tx postgresql.DBTX) error) (bool, error)
}
