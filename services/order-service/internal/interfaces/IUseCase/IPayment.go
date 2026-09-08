package IUseCase

import (
	"context"
	"kaffein/order-service/internal/dto"
)

type PaymentUseCase interface {
	Create(ctx context.Context, args dto.CreatePaymentDto) (string, error)
	Callback(ctx context.Context, args dto.MidtransCallbackDto) error
}
