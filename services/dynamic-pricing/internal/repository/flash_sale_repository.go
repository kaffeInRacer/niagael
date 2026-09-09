package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"kaffein/dynamic-pricing-service/internal/domain"
	"kaffein/dynamic-pricing-service/internal/dto"
	"kaffein/dynamic-pricing-service/internal/interfaces/IRepository"
	"kaffein/dynamic-pricing-service/pkg/postgresql"
	"kaffein/dynamic-pricing-service/utils/constants"
)

type flashSaleRepository struct {
	db postgresql.DBTX
}

func NewFlashSaleRepository(store *postgresql.Store) IRepository.FlashSaleRepository {
	return &flashSaleRepository{
		db: store,
	}
}

func (r *flashSaleRepository) List(ctx context.Context, params dto.ListFlashSaleParams) ([]domain.FlashSale, error) {
	const query = `
		SELECT id, name, product_id, variant_id, discount_percent, stock, max_per_user, start_time, end_time, is_active, created_at, updated_at
		FROM flash_sale
		WHERE (NULLIF($1::text, '') IS NULL OR name ILIKE '%' || $1::text || '%')
		AND ($2::boolean IS NULL OR is_active = $2::boolean)
		AND (NULLIF($3::text, '') IS NULL OR product_id = NULLIF($3::text, '')::uuid)
		AND (NULLIF($4::text, '') IS NULL OR discount_percent >= NULLIF($4::text, '')::numeric)
		AND (NULLIF($5::text, '') IS NULL OR discount_percent <= NULLIF($5::text, '')::numeric)
		AND (NULLIF($11::text, '') IS NULL OR name = $11::text)
		AND (NOT $6::boolean OR (stock > 0 AND NOW() BETWEEN start_time AND end_time))
		AND deleted_at IS NULL
		ORDER BY
		  CASE WHEN $7::text = 'name'       AND $8::text = 'asc'  THEN name       END ASC,
		  CASE WHEN $7::text = 'start_time' AND $8::text = 'asc'  THEN start_time END ASC,
		  CASE WHEN $7::text = 'end_time'   AND $8::text = 'asc'  THEN end_time   END ASC,
		  CASE WHEN $7::text = 'created_at' AND $8::text = 'asc'  THEN created_at END ASC,
		  CASE WHEN $7::text = 'name'       AND $8::text = 'desc' THEN name       END DESC,
		  CASE WHEN $7::text = 'start_time' AND $8::text = 'desc' THEN start_time END DESC,
		  CASE WHEN $7::text = 'end_time'   AND $8::text = 'desc' THEN end_time   END DESC,
		  CASE WHEN $7::text = 'created_at' AND $8::text = 'desc' THEN created_at END DESC,
		created_at DESC
		LIMIT $9::int
		OFFSET $10::int
	`
	rows, err := r.db.Query(ctx, query,
		params.Search,
		params.IsActive,
		params.ProductId,
		params.MinDiscountPercent,
		params.MaxDiscountPercent,
		params.CurrentOnly,
		params.OrderBy,
		params.OrderDir,
		params.PageSize,
		params.PageOffset,
		params.Name,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var flashSales []domain.FlashSale
	for rows.Next() {
		var fs domain.FlashSale
		if err := rows.Scan(
			&fs.Id,
			&fs.Name,
			&fs.ProductId,
			&fs.VariantId,
			&fs.DiscountPercent,
			&fs.Stock,
			&fs.MaxPerUser,
			&fs.StartTime,
			&fs.EndTime,
			&fs.IsActive,
			&fs.CreatedAt,
			&fs.UpdatedAt,
		); err != nil {
			return nil, err
		}
		flashSales = append(flashSales, fs)
	}

	return flashSales, rows.Err()
}

func (r *flashSaleRepository) ListSessions(ctx context.Context, params dto.ListFlashSaleParams) ([]dto.FlashSaleSession, int64, error) {
	const countQuery = `
		SELECT COUNT(DISTINCT name) FROM flash_sale WHERE deleted_at IS NULL
	`
	var count int64
	if err := r.db.QueryRow(ctx, countQuery).Scan(&count); err != nil {
		return nil, 0, err
	}

	const query = `
		SELECT name,
		       COUNT(*) AS item_count,
		       SUM(stock) AS total_stock,
		       MIN(start_time) AS start_time,
		       MAX(end_time) AS end_time,
		       BOOL_OR(is_active) AS is_active
		FROM flash_sale
		WHERE deleted_at IS NULL
		AND (NULLIF($1::text, '') IS NULL OR name ILIKE '%' || $1::text || '%')
		GROUP BY name, start_time, end_time
		ORDER BY MIN(created_at) DESC
		LIMIT $2::int OFFSET $3::int
	`
	rows, err := r.db.Query(ctx, query, params.Search, params.PageSize, params.PageOffset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var sessions []dto.FlashSaleSession
	for rows.Next() {
		var session dto.FlashSaleSession
		if err := rows.Scan(&session.Name, &session.ItemCount, &session.TotalStock, &session.StartTime, &session.EndTime, &session.IsActive); err != nil {
			return nil, 0, err
		}
		sessions = append(sessions, session)
	}
	return sessions, count, rows.Err()
}

func (r *flashSaleRepository) ListCount(ctx context.Context, params dto.ListFlashSaleParams) (int64, error) {
	const query = `
		SELECT COUNT(*)
		FROM flash_sale
		WHERE (NULLIF($1::text, '') IS NULL OR name ILIKE '%' || $1::text || '%')
		AND ($2::boolean IS NULL OR is_active = $2::boolean)
		AND (NULLIF($3::text, '') IS NULL OR product_id = NULLIF($3::text, '')::uuid)
		AND (NULLIF($4::text, '') IS NULL OR discount_percent >= NULLIF($4::text, '')::numeric)
		AND (NULLIF($5::text, '') IS NULL OR discount_percent <= NULLIF($5::text, '')::numeric)
		AND (NOT $6::boolean OR (stock > 0 AND NOW() BETWEEN start_time AND end_time))
		AND (NULLIF($7::text, '') IS NULL OR name = $7::text)
		AND deleted_at IS NULL
	`

	var count int64
	err := r.db.QueryRow(ctx, query,
		params.Search,
		params.IsActive,
		params.ProductId,
		params.MinDiscountPercent,
		params.MaxDiscountPercent,
		params.CurrentOnly,
		params.Name,
	).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (r *flashSaleRepository) Create(ctx context.Context, args dto.CreateFlashSaleDto) error {
	const query = `
		INSERT INTO flash_sale (id, name, product_id, variant_id, discount_percent, stock, max_per_user, start_time, end_time, is_active, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW())
	`
	_, err := r.db.Exec(ctx, query, args.Id, args.Name, args.ProductId, args.VariantId, args.DiscountPercent, args.Stock, args.MaxPerUser, args.StartTime, args.EndTime, args.IsActive)
	if err != nil {
		return err
	}

	return nil
}

func (r *flashSaleRepository) CreateBulk(ctx context.Context, args dto.CreateFlashSaleBulkDto) error {
	rows := make([][]any, len(args.Items))
	for i, item := range args.Items {
		rows[i] = []any{
			item.Id,
			args.Name,
			item.ProductId,
			item.VariantId,
			item.DiscountPercent,
			item.Stock,
			item.MaxPerUser,
			args.StartTime,
			args.EndTime,
			args.IsActive,
			time.Now(),
		}
	}

	columns := []string{"id", "name", "product_id", "variant_id", "discount_percent", "stock", "max_per_user", "start_time", "end_time", "is_active", "created_at"}

	_, err := r.db.CopyFrom(ctx,
		pgx.Identifier{"flash_sale"},
		columns,
		pgx.CopyFromRows(rows),
	)
	if err != nil {
		return err
	}

	return nil
}

func (r *flashSaleRepository) Update(ctx context.Context, args dto.UpdateFlashSaleDto) error {
	const query = `
		UPDATE flash_sale
		SET name = $2, product_id = $3, variant_id = $4, discount_percent = $5, stock = $6, max_per_user = $7, start_time = $8, end_time = $9, is_active = $10, updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`
	cmdTag, err := r.db.Exec(ctx, query, args.Id, args.Name, args.ProductId, args.VariantId, args.DiscountPercent, args.Stock, args.MaxPerUser, args.StartTime, args.EndTime, args.IsActive)
	if err != nil {
		return err
	}

	if cmdTag.RowsAffected() == 0 {
		return errors.New(constants.ErrFlashSaleNotFound)
	}

	return nil
}

func (r *flashSaleRepository) Delete(ctx context.Context, id string) error {
	const query = `
		UPDATE flash_sale
		SET deleted_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`

	cmdTag, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	if cmdTag.RowsAffected() == 0 {
		return errors.New(constants.ErrFlashSaleNotFound)
	}

	return nil
}

func (r *flashSaleRepository) ReadById(ctx context.Context, id string) (*domain.FlashSale, error) {
	const query = `
		SELECT id, name, product_id, variant_id, discount_percent, stock, max_per_user, start_time, end_time, is_active, created_at, updated_at
		FROM flash_sale
		WHERE id = $1 AND deleted_at IS NULL
	`

	var fs domain.FlashSale
	err := r.db.QueryRow(ctx, query, id).Scan(
		&fs.Id,
		&fs.Name,
		&fs.ProductId,
		&fs.VariantId,
		&fs.DiscountPercent,
		&fs.Stock,
		&fs.MaxPerUser,
		&fs.StartTime,
		&fs.EndTime,
		&fs.IsActive,
		&fs.CreatedAt,
		&fs.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &fs, nil
}

func (r *flashSaleRepository) ReadByProductId(ctx context.Context, productId string) (*domain.FlashSale, error) {
	const query = `
		SELECT id, name, product_id, variant_id, discount_percent, stock, max_per_user, start_time, end_time, is_active, created_at, updated_at
		FROM flash_sale
		WHERE product_id = $1
		  AND variant_id IS NULL
		  AND is_active = true
		  AND stock > 0
		  AND NOW() BETWEEN start_time AND end_time
		  AND deleted_at IS NULL
		ORDER BY start_time DESC
		LIMIT 1
	`

	var fs domain.FlashSale
	err := r.db.QueryRow(ctx, query, productId).Scan(
		&fs.Id,
		&fs.Name,
		&fs.ProductId,
		&fs.VariantId,
		&fs.DiscountPercent,
		&fs.Stock,
		&fs.MaxPerUser,
		&fs.StartTime,
		&fs.EndTime,
		&fs.IsActive,
		&fs.CreatedAt,
		&fs.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &fs, nil
}

func (r *flashSaleRepository) ReadByVariantId(ctx context.Context, productId string, variantId string) (*domain.FlashSale, error) {
	const query = `
		SELECT id, name, product_id, variant_id, discount_percent, stock, max_per_user, start_time, end_time, is_active, created_at, updated_at
		FROM flash_sale
		WHERE product_id = $1
		  AND variant_id = $2
		  AND is_active = true
		  AND stock > 0
		  AND NOW() BETWEEN start_time AND end_time
		  AND deleted_at IS NULL
		ORDER BY start_time DESC
		LIMIT 1
	`

	var fs domain.FlashSale
	err := r.db.QueryRow(ctx, query, productId, variantId).Scan(
		&fs.Id,
		&fs.Name,
		&fs.ProductId,
		&fs.VariantId,
		&fs.DiscountPercent,
		&fs.Stock,
		&fs.MaxPerUser,
		&fs.StartTime,
		&fs.EndTime,
		&fs.IsActive,
		&fs.CreatedAt,
		&fs.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &fs, nil
}

func (r *flashSaleRepository) ReadByProductIds(ctx context.Context, productIds []string) ([]domain.FlashSale, error) {
	if len(productIds) == 0 {
		return nil, nil
	}

	const query = `
		SELECT id, name, product_id, variant_id, discount_percent, stock, max_per_user, start_time, end_time, is_active, created_at, updated_at
		FROM flash_sale
		WHERE product_id = ANY($1)
		  AND variant_id IS NULL
		  AND is_active = true
		  AND stock > 0
		  AND NOW() BETWEEN start_time AND end_time
		  AND deleted_at IS NULL
	`

	rows, err := r.db.Query(ctx, query, productIds)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var flashSales []domain.FlashSale
	for rows.Next() {
		var fs domain.FlashSale
		if err := rows.Scan(
			&fs.Id,
			&fs.Name,
			&fs.ProductId,
			&fs.VariantId,
			&fs.DiscountPercent,
			&fs.Stock,
			&fs.MaxPerUser,
			&fs.StartTime,
			&fs.EndTime,
			&fs.IsActive,
			&fs.CreatedAt,
			&fs.UpdatedAt,
		); err != nil {
			return nil, err
		}
		flashSales = append(flashSales, fs)
	}

	return flashSales, rows.Err()
}

func (r *flashSaleRepository) ReadByVariantIds(ctx context.Context, items []dto.VariantIdPair) ([]domain.FlashSale, error) {
	if len(items) == 0 {
		return nil, nil
	}

	const query = `
		SELECT id, name, product_id, variant_id, discount_percent, stock, max_per_user, start_time, end_time, is_active, created_at, updated_at
		FROM flash_sale
		WHERE (product_id, variant_id) IN (SELECT unnest($1::uuid[]), unnest($2::uuid[]))
		  AND is_active = true
		  AND stock > 0
		  AND NOW() BETWEEN start_time AND end_time
		  AND deleted_at IS NULL
	`

	productIds := make([]string, len(items))
	variantIds := make([]string, len(items))
	for i, item := range items {
		productIds[i] = item.ProductId
		variantIds[i] = item.VariantId
	}

	rows, err := r.db.Query(ctx, query, productIds, variantIds)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var flashSales []domain.FlashSale
	for rows.Next() {
		var fs domain.FlashSale
		if err := rows.Scan(
			&fs.Id,
			&fs.Name,
			&fs.ProductId,
			&fs.VariantId,
			&fs.DiscountPercent,
			&fs.Stock,
			&fs.MaxPerUser,
			&fs.StartTime,
			&fs.EndTime,
			&fs.IsActive,
			&fs.CreatedAt,
			&fs.UpdatedAt,
		); err != nil {
			return nil, err
		}
		flashSales = append(flashSales, fs)
	}

	return flashSales, rows.Err()
}

func (r *flashSaleRepository) DecrementStock(ctx context.Context, id string, quantity int) error {
	const query = `
		UPDATE flash_sale
		SET stock = stock - $2, updated_at = NOW()
		WHERE id = $1 AND stock >= $2 AND deleted_at IS NULL
	`
	cmdTag, err := r.db.Exec(ctx, query, id, quantity)
	if err != nil {
		return err
	}
	if cmdTag.RowsAffected() == 0 {
		return errors.New(constants.ErrFlashSaleStockInsufficient)
	}
	return nil
}
