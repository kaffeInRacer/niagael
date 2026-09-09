package repository

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"kaffein/dynamic-pricing-service/internal/domain"
	"kaffein/dynamic-pricing-service/internal/dto"
	"kaffein/dynamic-pricing-service/internal/interfaces/IRepository"
	"kaffein/dynamic-pricing-service/pkg/postgresql"
	"kaffein/dynamic-pricing-service/utils/constants"
)

type promoRepository struct {
	db postgresql.DBTX
}

func NewPromoRepository(store *postgresql.Store) IRepository.PromoRepository {
	return &promoRepository{
		db: store,
	}
}

func (r *promoRepository) List(ctx context.Context, params dto.ListPromoParams) ([]domain.Promo, error) {
	const query = `
		SELECT id, name, code, description, discount_type, discount_value, min_purchase, max_discount,
		       quantity, used_count, max_usage_per_user, can_combine_flash_sale, start_date, end_date, is_active, created_at, updated_at
		FROM promo
		WHERE (NULLIF($1::text, '') IS NULL OR name ILIKE '%' || $1::text || '%' OR code ILIKE '%' || $1::text || '%')
		AND ($2::boolean IS NULL OR is_active = $2::boolean)
		AND (NOT $3::boolean OR (NOW() BETWEEN start_date AND end_date AND (quantity = 0 OR used_count < quantity)))
		AND deleted_at IS NULL
		ORDER BY
		  CASE WHEN $4::text = 'name'       AND $5::text = 'asc'  THEN name       END ASC,
		  CASE WHEN $4::text = 'code'       AND $5::text = 'asc'  THEN code       END ASC,
		  CASE WHEN $4::text = 'start_date' AND $5::text = 'asc'  THEN start_date END ASC,
		  CASE WHEN $4::text = 'created_at' AND $5::text = 'asc'  THEN created_at END ASC,
		  CASE WHEN $4::text = 'name'       AND $5::text = 'desc' THEN name       END DESC,
		  CASE WHEN $4::text = 'code'       AND $5::text = 'desc' THEN code       END DESC,
		  CASE WHEN $4::text = 'start_date' AND $5::text = 'desc' THEN start_date END DESC,
		  CASE WHEN $4::text = 'created_at' AND $5::text = 'desc' THEN created_at END DESC,
		created_at DESC
		LIMIT $6::int
		OFFSET $7::int
	`
	rows, err := r.db.Query(ctx, query,
		params.Search,
		params.IsActive,
		params.CurrentOnly,
		params.OrderBy,
		params.OrderDir,
		params.PageSize,
		params.PageOffset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var promos []domain.Promo
	for rows.Next() {
		var p domain.Promo
		if err := rows.Scan(
			&p.Id,
			&p.Name,
			&p.Code,
			&p.Description,
			&p.DiscountType,
			&p.DiscountValue,
			&p.MinPurchase,
			&p.MaxDiscount,
			&p.Quantity,
			&p.UsedCount,
			&p.MaxUsagePerUser,
			&p.CanCombineFlashSale,
			&p.StartDate,
			&p.EndDate,
			&p.IsActive,
			&p.CreatedAt,
			&p.UpdatedAt,
		); err != nil {
			return nil, err
		}
		promos = append(promos, p)
	}

	return promos, rows.Err()
}

func (r *promoRepository) ListCount(ctx context.Context, params dto.ListPromoParams) (int64, error) {
	const query = `
		SELECT COUNT(*)
		FROM promo
		WHERE (NULLIF($1::text, '') IS NULL OR name ILIKE '%' || $1::text || '%' OR code ILIKE '%' || $1::text || '%')
		AND ($2::boolean IS NULL OR is_active = $2::boolean)
		AND (NOT $3::boolean OR (NOW() BETWEEN start_date AND end_date AND (quantity = 0 OR used_count < quantity)))
		AND deleted_at IS NULL
	`

	var count int64
	err := r.db.QueryRow(ctx, query,
		params.Search,
		params.IsActive,
		params.CurrentOnly,
	).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (r *promoRepository) Create(ctx context.Context, args dto.CreatePromoDto) error {
	const query = `
		INSERT INTO promo (id, name, code, description, discount_type, discount_value, min_purchase, max_discount,
		                   quantity, max_usage_per_user, can_combine_flash_sale, start_date, end_date, is_active, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, NOW())
	`
	_, err := r.db.Exec(ctx, query,
		args.Id, args.Name, args.Code, args.Description, args.DiscountType, args.DiscountValue,
		args.MinPurchase, args.MaxDiscount, args.Quantity, args.MaxUsagePerUser, args.CanCombineFlashSale,
		args.StartDate, args.EndDate, args.IsActive)
	if err != nil {
		return err
	}

	return nil
}

func (r *promoRepository) Update(ctx context.Context, args dto.UpdatePromoDto) error {
	const query = `
		UPDATE promo
		SET name = $2, code = $3, description = $4, discount_type = $5, discount_value = $6, 
		    min_purchase = $7, max_discount = $8, quantity = $9, max_usage_per_user = $10,
		    can_combine_flash_sale = $11, start_date = $12, end_date = $13, is_active = $14, updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`
	cmdTag, err := r.db.Exec(ctx, query,
		args.Id, args.Name, args.Code, args.Description, args.DiscountType, args.DiscountValue,
		args.MinPurchase, args.MaxDiscount, args.Quantity, args.MaxUsagePerUser, args.CanCombineFlashSale,
		args.StartDate, args.EndDate, args.IsActive)
	if err != nil {
		return err
	}

	if cmdTag.RowsAffected() == 0 {
		return errors.New(constants.ErrPromoNotFound)
	}

	return nil
}

func (r *promoRepository) Delete(ctx context.Context, id string) error {
	const query = `
		UPDATE promo
		SET deleted_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`

	cmdTag, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	if cmdTag.RowsAffected() == 0 {
		return errors.New(constants.ErrPromoNotFound)
	}

	return nil
}

func (r *promoRepository) ReadById(ctx context.Context, id string) (*domain.Promo, error) {
	const query = `
		SELECT id, name, code, description, discount_type, discount_value, min_purchase, max_discount,
		       quantity, used_count, max_usage_per_user, can_combine_flash_sale, start_date, end_date, is_active, created_at, updated_at
		FROM promo
		WHERE id = $1 AND deleted_at IS NULL
	`

	var p domain.Promo
	err := r.db.QueryRow(ctx, query, id).Scan(
		&p.Id,
		&p.Name,
		&p.Code,
		&p.Description,
		&p.DiscountType,
		&p.DiscountValue,
		&p.MinPurchase,
		&p.MaxDiscount,
		&p.Quantity,
		&p.UsedCount,
		&p.MaxUsagePerUser,
		&p.CanCombineFlashSale,
		&p.StartDate,
		&p.EndDate,
		&p.IsActive,
		&p.CreatedAt,
		&p.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &p, nil
}

func (r *promoRepository) ReadByCode(ctx context.Context, code string) (*domain.Promo, error) {
	const query = `
		SELECT id, name, code, description, discount_type, discount_value, min_purchase, max_discount,
		       quantity, used_count, max_usage_per_user, can_combine_flash_sale, start_date, end_date, is_active, created_at, updated_at
		FROM promo
		WHERE code = $1 AND deleted_at IS NULL
	`

	var p domain.Promo
	err := r.db.QueryRow(ctx, query, code).Scan(
		&p.Id,
		&p.Name,
		&p.Code,
		&p.Description,
		&p.DiscountType,
		&p.DiscountValue,
		&p.MinPurchase,
		&p.MaxDiscount,
		&p.Quantity,
		&p.UsedCount,
		&p.MaxUsagePerUser,
		&p.CanCombineFlashSale,
		&p.StartDate,
		&p.EndDate,
		&p.IsActive,
		&p.CreatedAt,
		&p.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &p, nil
}

func (r *promoRepository) IncrementUsage(ctx context.Context, id string) error {
	const query = `
		UPDATE promo
		SET used_count = used_count + 1, updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`
	_, err := r.db.Exec(ctx, query, id)
	return err
}
