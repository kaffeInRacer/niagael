package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"kaffein/product-service/internal/domain"
)

type FlashSaleProjectionRepository struct {
	db *pgxpool.Pool
}

func NewFlashSaleProjectionRepository(pool *pgxpool.Pool) *FlashSaleProjectionRepository {
	return &FlashSaleProjectionRepository{db: pool}
}

func (r *FlashSaleProjectionRepository) Upsert(ctx context.Context, p domain.FlashSaleProjection) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO flash_sale_projection (product_id, variant_id, name, discount_percent, start_time, end_time, is_active, updated_at)
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

func (r *FlashSaleProjectionRepository) Delete(ctx context.Context, productID, variantID string) error {
	_, err := r.db.Exec(ctx, `DELETE FROM flash_sale_projection WHERE product_id = $1 AND variant_id = $2`, productID, variantID)
	return err
}

func (r *FlashSaleProjectionRepository) Count(ctx context.Context) (int, error) {
	var count int
	err := r.db.QueryRow(ctx, `SELECT count(*) FROM flash_sale_projection`).Scan(&count)
	return count, err
}

func (r *FlashSaleProjectionRepository) ListActive(ctx context.Context) ([]domain.FlashSaleProjection, error) {
	rows, err := r.db.Query(ctx, `
		SELECT product_id, variant_id, name, discount_percent, start_time, end_time, is_active
		FROM flash_sale_projection
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
