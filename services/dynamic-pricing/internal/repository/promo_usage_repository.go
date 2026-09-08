package repository

import (
	"context"
	"kaffein/dynamic-pricing-service/internal/domain"
	"kaffein/dynamic-pricing-service/internal/interfaces/IRepository"
	"kaffein/dynamic-pricing-service/pkg/postgresql"
)

type promoUsageRepository struct {
	db postgresql.DBTX
}

func NewPromoUsageRepository(store *postgresql.Store) IRepository.PromoUsageRepository {
	return &promoUsageRepository{
		db: store,
	}
}

func (r *promoUsageRepository) GetUsage(ctx context.Context, promoId string, userId string) (*domain.PromoUsage, error) {
	const query = `
		SELECT id, promo_id, user_id, quantity, created_at, updated_at
		FROM promo_usage
		WHERE promo_id = $1 AND user_id = $2
	`

	var usage domain.PromoUsage
	err := r.db.QueryRow(ctx, query, promoId, userId).Scan(
		&usage.Id,
		&usage.PromoId,
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

func (r *promoUsageRepository) IncrementUsage(ctx context.Context, promoId string, userId string, quantity int) error {
	const query = `
		INSERT INTO promo_usage (id, promo_id, user_id, quantity, created_at)
		VALUES (gen_random_uuid(), $1, $2, $3, NOW())
		ON CONFLICT (promo_id, user_id) 
		DO UPDATE SET quantity = promo_usage.quantity + $3, updated_at = NOW()
	`
	_, err := r.db.Exec(ctx, query, promoId, userId, quantity)
	return err
}
