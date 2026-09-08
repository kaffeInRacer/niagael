package IRepository

import (
	"context"

	"kaffein/order-service/internal/domain"
)

type CartRepository interface {
	ReadByUserId(ctx context.Context, userId string) (*domain.Cart, error)
	AddItem(ctx context.Context, userId string, cartId string, item domain.CartItem, maxQuantity int) error
	UpdateItem(ctx context.Context, itemId string, userId string, quantity int, maxQuantity int) error
	DeleteItem(ctx context.Context, itemId string, userId string) error
	Clear(ctx context.Context, userId string) error
}
