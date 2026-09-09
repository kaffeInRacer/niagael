package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type IdempotencyRepository struct {
	db *pgxpool.Pool
}

func NewIdempotencyRepository(pool *pgxpool.Pool) *IdempotencyRepository {
	return &IdempotencyRepository{db: pool}
}

// ClaimOrderKey atomically claims an idempotency key for the given order id.
// Returns (winner order id, true) when this call won the claim, or
// (existing order id, false) when the key was already claimed.
func (r *IdempotencyRepository) ClaimOrderKey(ctx context.Context, key, orderID string) (string, bool, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return "", false, err
	}
	defer tx.Rollback(ctx)

	var winner string
	err = tx.QueryRow(ctx, `
		INSERT INTO order_request_keys (key, order_id)
		VALUES ($1, $2)
		ON CONFLICT (key) DO NOTHING
		RETURNING order_id::text
	`, key, orderID).Scan(&winner)
	if err == nil {
		if err := tx.Commit(ctx); err != nil {
			return "", false, err
		}
		return winner, true, nil
	}
	if err != pgx.ErrNoRows {
		return "", false, err
	}

	// Key already exists: read its owner under the same transaction for a
	// consistent view.
	err = tx.QueryRow(ctx, `SELECT order_id::text FROM order_request_keys WHERE key = $1 FOR UPDATE`, key).Scan(&winner)
	if err != nil {
		return "", false, err
	}
	return winner, false, nil
}

func (r *IdempotencyRepository) ReleaseOrderKey(ctx context.Context, key string) error {
	_, err := r.db.Exec(ctx, `DELETE FROM order_request_keys WHERE key = $1`, key)
	return err
}
