package repository

import (
	"context"
	"errors"
	"fmt"

	"kaffein/product-service/internal/domain"
	"kaffein/product-service/internal/interfaces/IRepository"
	"kaffein/product-service/pkg/postgresql"
	"kaffein/product-service/utils/constants"
)

type stockRepository struct {
	store *postgresql.Store
}

func NewStockRepository(store *postgresql.Store) IRepository.StockRepository {
	return &stockRepository{
		store: store,
	}
}

func (r *stockRepository) ReserveWithTx(ctx context.Context, orderID string, items []domain.StockItem) error {
	return r.mutateWithTx(ctx, orderID, items, domain.StockReserve)
}

func (r *stockRepository) ConfirmWithTx(ctx context.Context, orderID string, items []domain.StockItem) error {
	return r.mutateWithTx(ctx, orderID, items, domain.StockConfirm)
}

func (r *stockRepository) ReleaseWithTx(ctx context.Context, orderID string, items []domain.StockItem) error {
	return r.mutateWithTx(ctx, orderID, items, domain.StockRelease)
}

func (r *stockRepository) mutateWithTx(ctx context.Context, orderID string, rawItems []domain.StockItem, operation domain.StockOperation) error {
	items, fingerprint, payload, err := domain.CanonicalizeStockItems(rawItems)
	if err != nil {
		return err
	}

	return r.store.ExecTx(ctx, func(tx postgresql.DBTX) error {
		const upsertQuery = `
			INSERT INTO stock_reservation (order_id, request_hash, items, status, created_at, updated_at)
			VALUES ($1, $2, $3, $4, NOW(), NOW())
			ON CONFLICT (order_id) DO NOTHING
		`
		command, err := tx.Exec(ctx, upsertQuery, orderID, fingerprint, payload, domain.InitialStockState(operation))
		if err != nil {
			return err
		}

		const lockQuery = `
			SELECT request_hash, status FROM stock_reservation WHERE order_id = $1 FOR UPDATE
		`
		var storedHash, state string
		if err := tx.QueryRow(ctx, lockQuery, orderID).Scan(&storedHash, &state); err != nil {
			return err
		}
		if storedHash != fingerprint {
			return fmt.Errorf("%s: order ID already has different items", constants.ErrInvalidStockRequest)
		}

		isNew := command.RowsAffected() == 1
		skip, nextState, err := domain.StockTransition(operation, isNew, state)
		if err != nil {
			return err
		}
		if skip {
			return nil
		}

		if err := mutateStockItems(ctx, tx, items, string(operation)); err != nil {
			return err
		}

		if nextState == "" {
			return nil
		}

		const updateStatusQuery = `
			UPDATE stock_reservation SET status = $2, updated_at = NOW() WHERE order_id = $1
		`
		_, err = tx.Exec(ctx, updateStatusQuery, orderID, nextState)
		return err
	})
}

func mutateStockItems(ctx context.Context, tx postgresql.DBTX, items []domain.CanonicalStockItem, operation string) error {
	for _, item := range items {
		var query string
		args := []any{item.Quantity, item.ProductID}

		if item.VariantID == "" {
			switch operation {
			case "reserve":
				query = `UPDATE product SET stock_reserved = stock_reserved + $1 WHERE id = $2 AND stock - stock_reserved >= $1 AND deleted_at IS NULL`
			case "confirm":
				query = `UPDATE product SET stock = stock - $1, stock_reserved = stock_reserved - $1 WHERE id = $2 AND stock_reserved >= $1 AND deleted_at IS NULL`
			case "release":
				query = `UPDATE product SET stock_reserved = stock_reserved - $1 WHERE id = $2 AND stock_reserved >= $1 AND deleted_at IS NULL`
			}
		} else {
			args = append(args, item.VariantID)
			switch operation {
			case "reserve":
				query = `UPDATE product_variant SET stock_reserved = stock_reserved + $1 WHERE product_id = $2 AND id = $3 AND stock - stock_reserved >= $1 AND deleted_at IS NULL`
			case "confirm":
				query = `UPDATE product_variant SET stock = stock - $1, stock_reserved = stock_reserved - $1 WHERE product_id = $2 AND id = $3 AND stock_reserved >= $1 AND deleted_at IS NULL`
			case "release":
				query = `UPDATE product_variant SET stock_reserved = stock_reserved - $1 WHERE product_id = $2 AND id = $3 AND stock_reserved >= $1 AND deleted_at IS NULL`
			}
		}

		command, err := tx.Exec(ctx, query, args...)
		if err != nil {
			return err
		}

		if command.RowsAffected() != 1 {
			return errors.New(constants.ErrInsufficientStock)
		}
	}
	return nil
}
