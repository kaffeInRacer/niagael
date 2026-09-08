package repository

import (
	"context"
	"kaffein/dynamic-pricing-service/internal/domain"
	"kaffein/dynamic-pricing-service/internal/interfaces/IRepository"
	"kaffein/dynamic-pricing-service/pkg/postgresql"
)

type flashSaleUsageRepository struct {
	db postgresql.DBTX
}

func NewFlashSaleUsageRepository(store *postgresql.Store) IRepository.FlashSaleUsageRepository {
	return &flashSaleUsageRepository{
		db: store,
	}
}

func (r *flashSaleUsageRepository) GetUsage(ctx context.Context, flashSaleId string, userId string) (*domain.FlashSaleUsage, error) {
	const query = `
		SELECT id, flash_sale_id, user_id, quantity, created_at, updated_at
		FROM flash_sale_usage
		WHERE flash_sale_id = $1 AND user_id = $2
	`

	var usage domain.FlashSaleUsage
	err := r.db.QueryRow(ctx, query, flashSaleId, userId).Scan(
		&usage.Id,
		&usage.FlashSaleId,
		&usage.UserId,
		&usage.Quantity,
		&usage.CreatedAt,
		&usage.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &usage, nil
}

func (r *flashSaleUsageRepository) IncrementUsage(ctx context.Context, flashSaleId string, userId string, quantity int) error {
	const query = `
		INSERT INTO flash_sale_usage (id, flash_sale_id, user_id, quantity, created_at)
		VALUES (gen_random_uuid(), $1, $2, $3, NOW())
		ON CONFLICT (flash_sale_id, user_id) 
		DO UPDATE SET quantity = flash_sale_usage.quantity + $3, updated_at = NOW()
	`
	_, err := r.db.Exec(ctx, query, flashSaleId, userId, quantity)
	return err
}
