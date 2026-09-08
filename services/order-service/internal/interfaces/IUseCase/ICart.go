package IUseCase

import (
	"context"

	"kaffein/order-service/internal/dto"
)

type CartUseCase interface {
	ReadByUserId(ctx context.Context, userId string) (*dto.CartResponse, error)
	AddItem(ctx context.Context, args dto.AddCartItemDto) (*dto.CartResponse, error)
	UpdateItem(ctx context.Context, itemId string, args dto.UpdateCartItemDto) (*dto.CartResponse, error)
	DeleteItem(ctx context.Context, itemId string, userId string) error
	Clear(ctx context.Context, userId string) error
}
