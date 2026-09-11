package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"kaffein/product-service/internal/domain"
)

type FlashSaleSnapshotRepository struct {
	db *pgxpool.Pool
}

func NewFlashSaleSnapshotRepository(pool *pgxpool.Pool) *FlashSaleSnapshotRepository {
	return &FlashSaleSnapshotRepository{db: pool}
}

func (r *FlashSaleSnapshotRepository) Upsert(ctx context.Context, p domain.FlashSaleProjection) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO flash_sale_snapshot (product_id, variant_id, name, discount_percent, start_time, end_time, is_active, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
		ON CONFLICT (product_id, variant_id) DO UPDATE SET
			name = EXCLUDED.name,
			discount_percent = EXCLUDED.discount_percent,
			start_time = EXCLUDED.start_time,
			end_time = EXCLUDED.end_time,
			is_active = EXCLUDED.is_active,
			updated_at = NOW()
	`, p.ProductID, p.VariantID, p.Name, p.DiscountPercent, p.StartTime, p.EndTime, p.IsActive)
	return err
}

func (r *FlashSaleSnapshotRepository) Delete(ctx context.Context, productID, variantID string) error {
	_, err := r.db.Exec(ctx, `DELETE FROM flash_sale_snapshot WHERE product_id = $1 AND variant_id = $2`, productID, variantID)
	return err
}

func (r *FlashSaleSnapshotRepository) Count(ctx context.Context) (int, error) {
	var count int
	err := r.db.QueryRow(ctx, `SELECT count(*) FROM flash_sale_snapshot`).Scan(&count)
	return count, err
}

func (r *FlashSaleSnapshotRepository) ListActive(ctx context.Context) ([]domain.FlashSaleProjection, error) {
	rows, err := r.db.Query(ctx, `
		SELECT product_id, variant_id, name, discount_percent, start_time, end_time, is_active
		FROM flash_sale_snapshot
		WHERE is_active = true AND start_time <= NOW() AND end_time >= NOW()
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []domain.FlashSaleProjection
	for rows.Next() {
		var p domain.FlashSaleProjection
		if err := rows.Scan(&p.ProductID, &p.VariantID, &p.Name, &p.DiscountPercent, &p.StartTime, &p.EndTime, &p.IsActive); err != nil {
			return nil, err
		}
		items = append(items, p)
	}
	return items, rows.Err()
}
