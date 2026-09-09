package IRepository

import (
	"context"
	"kaffein/product-service/internal/domain"
)

type StockRepository interface {
	ReserveWithTx(context.Context, string, []domain.StockItem) error
	ConfirmWithTx(context.Context, string, []domain.StockItem) error
	ReleaseWithTx(context.Context, string, []domain.StockItem) error
}
