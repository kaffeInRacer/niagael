package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type BuyersRepository struct {
	db *pgxpool.Pool
}

func NewBuyersRepository(pool *pgxpool.Pool) *BuyersRepository {
	return &BuyersRepository{db: pool}
}

func (r *BuyersRepository) Upsert(ctx context.Context, id, email string) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO buyers (id, email, updated_at)
		VALUES ($1::uuid, $2, NOW())
		ON CONFLICT (id) DO UPDATE SET email = EXCLUDED.email, updated_at = NOW()
	`, id, email)
	return err
}

func (r *BuyersRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, `DELETE FROM buyers WHERE id = $1::uuid`, id)
	return err
}

func (r *BuyersRepository) Snapshot(ctx context.Context, authPool *pgxpool.Pool) error {
	rows, err := authPool.Query(ctx, `SELECT id, email FROM users WHERE deleted_at IS NULL`)
	if err != nil {
		return err
	}
	defer rows.Close()

	batch := &pgx.Batch{}
	for rows.Next() {
		var id, email string
		if err := rows.Scan(&id, &email); err != nil {
			return err
		}
		batch.Queue(`INSERT INTO buyers (id, email, updated_at)
			VALUES ($1::uuid, $2, NOW())
			ON CONFLICT (id) DO UPDATE SET email = EXCLUDED.email, updated_at = NOW()`, id, email)
	}
	if err := rows.Err(); err != nil {
		return err
	}

	conn, err := r.db.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	br := conn.SendBatch(ctx, batch)
	defer br.Close()

	for range batch.Len() {
		if _, err := br.Exec(); err != nil {
			return err
		}
	}
	return nil
}
