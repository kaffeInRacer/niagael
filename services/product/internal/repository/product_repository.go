package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"kaffein/product-service/internal/domain"
	"kaffein/product-service/internal/dto"
	"kaffein/product-service/internal/interfaces/IRepository"
	"kaffein/product-service/pkg/postgresql"
	"kaffein/product-service/utils/constants"
)

type productRepository struct {
	db postgresql.DBTX
}

func NewProductRepository(store *postgresql.Store) IRepository.ProductRepository {
	return &productRepository{
		db: store,
	}
}

func productOrderBy(column, direction string) (string, string) {
	allowed := map[string]string{
		"name":       "v.name",
		"price":      "v.price",
		"created_at": "v.created_at",
		"updated_at": "v.updated_at",
	}

	col, ok := allowed[column]
	if !ok {
		col = "v.created_at"
	}

	dir := "DESC"
	if strings.EqualFold(direction, "asc") {
		dir = "ASC"
	}

	orderBy := fmt.Sprintf("%s %s", col, dir)
	fallback := "v.created_at DESC"
	return orderBy, fallback
}

const productColumns = `
		v.id, v.category_id, v.category_name, v.category_active,
		v.name, v.slug, v.description, v.price, v.stock, v.stock_reserved,
		v.is_active, v.is_promo_excluded, v.created_at, v.updated_at
`

func (r *productRepository) List(ctx context.Context, params dto.ListProductParams) ([]domain.Product, error) {
	orderBy, fallbackOrderBy := productOrderBy(params.OrderBy, params.OrderDir)

	query := fmt.Sprintf(`
		SELECT %s
		FROM product_list_view v
		WHERE (NULLIF($1::text, '') IS NULL OR v.name ILIKE '%%' || $1::text || '%%')
		AND ($2::boolean IS NULL OR v.is_active = $2::boolean)
		AND ($3::boolean IS NULL OR v.is_promo_excluded = $3::boolean)
		AND (NULLIF($4::text, '') IS NULL OR v.category_id = NULLIF($4::text, '')::uuid)
		AND (NULLIF($5::text, '') IS NULL OR v.price >= NULLIF($5::text, '')::numeric)
		AND (NULLIF($6::text, '') IS NULL OR v.price <= NULLIF($6::text, '')::numeric)
		AND (NOT $7::boolean OR (v.category_active = true))
		ORDER BY %s, %s
		LIMIT $8::int
		OFFSET $9::int
	`, productColumns, orderBy, fallbackOrderBy)

	rows, err := r.db.Query(ctx, query,
		params.Search,
		params.IsActive,
		params.IsPromoExcluded,
		params.CategoryID,
		params.MinPrice,
		params.MaxPrice,
		params.PublicOnly,
		params.PageSize,
		params.PageOffset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []domain.Product
	for rows.Next() {
		var product domain.Product
		if err := rows.Scan(
			&product.Id,
			&product.CategoryId,
			&product.CategoryName,
			&product.CategoryActive,
			&product.Name,
			&product.Slug,
			&product.Description,
			&product.Price,
			&product.Stock,
			&product.StockReserved,
			&product.IsActive,
			&product.IsPromo,
			&product.CreatedAt,
			&product.UpdatedAt,
		); err != nil {
			return nil, err
		}
		products = append(products, product)
	}

	return products, rows.Err()
}

func (r *productRepository) ListCount(ctx context.Context, params dto.ListProductParams) (int64, error) {
	const query = `
		SELECT COUNT(*)
		FROM product_list_view v
		WHERE (NULLIF($1::text, '') IS NULL OR v.name ILIKE '%' || $1::text || '%')
		AND ($2::boolean IS NULL OR v.is_active = $2::boolean)
		AND ($3::boolean IS NULL OR v.is_promo_excluded = $3::boolean)
		AND (NULLIF($4::text, '') IS NULL OR v.category_id = NULLIF($4::text, '')::uuid)
		AND (NULLIF($5::text, '') IS NULL OR v.price >= NULLIF($5::text, '')::numeric)
		AND (NULLIF($6::text, '') IS NULL OR v.price <= NULLIF($6::text, '')::numeric)
		AND (NOT $7::boolean OR (v.category_active = true))
	`

	var count int64
	err := r.db.QueryRow(ctx, query,
		params.Search,
		params.IsActive,
		params.IsPromoExcluded,
		params.CategoryID,
		params.MinPrice,
		params.MaxPrice,
		params.PublicOnly,
	).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (r *productRepository) Create(ctx context.Context, args dto.CreateProductDto) error {
	const query = `
		INSERT INTO product (id, category_id, name, slug, description, price, stock, is_active, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())
	`
	_, err := r.db.Exec(ctx, query, args.Id, args.CategoryId, args.Name, args.Slug, args.Description, args.Price, args.Stock, args.IsActive)
	if err != nil {
		return err
	}

	return nil
}

func (r *productRepository) Update(ctx context.Context, args dto.UpdateProductDto) error {
	const query = `
		UPDATE product
		SET category_id = $2, name = $3, slug = $4, description = $5,
		    price = $6, stock = $7, is_active = $8, updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`
	cmdTag, err := r.db.Exec(ctx, query, args.Id, args.CategoryId, args.Name, args.Slug, args.Description, args.Price, args.Stock, args.IsActive)
	if err != nil {
		return err
	}

	if cmdTag.RowsAffected() == 0 {
		return errors.New(constants.ErrProductNotFound)
	}

	return nil
}

func (r *productRepository) Delete(ctx context.Context, id string) error {
	const query = `
		UPDATE product
		SET deleted_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`

	cmdTag, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	if cmdTag.RowsAffected() == 0 {
		return errors.New(constants.ErrProductNotFound)
	}

	return nil
}

func (r *productRepository) ReadById(ctx context.Context, id string) (*domain.Product, error) {
	const query = `
		SELECT %s
		FROM product_list_view v
		WHERE v.id = $1
	`

	var product domain.Product
	err := r.db.QueryRow(ctx, fmt.Sprintf(query, productColumns), id).Scan(
		&product.Id,
		&product.CategoryId,
		&product.CategoryName,
		&product.CategoryActive,
		&product.Name,
		&product.Slug,
		&product.Description,
		&product.Price,
		&product.Stock,
		&product.StockReserved,
		&product.IsActive,
		&product.IsPromo,
		&product.CreatedAt,
		&product.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &product, nil
}

func (r *productRepository) ReadBySlug(ctx context.Context, slug string) (*domain.Product, error) {
	const query = `
		SELECT %s
		FROM product_list_view v
		WHERE v.slug = $1
	`

	var product domain.Product
	err := r.db.QueryRow(ctx, fmt.Sprintf(query, productColumns), slug).Scan(
		&product.Id,
		&product.CategoryId,
		&product.CategoryName,
		&product.CategoryActive,
		&product.Name,
		&product.Slug,
		&product.Description,
		&product.Price,
		&product.Stock,
		&product.StockReserved,
		&product.IsActive,
		&product.IsPromo,
		&product.CreatedAt,
		&product.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &product, nil
}

func (r *productRepository) BatchById(ctx context.Context, ids []string) ([]domain.Product, error) {
	const query = `
		SELECT %s
		FROM product_list_view v
		WHERE v.id = ANY($1) AND v.is_active = true
	`

	rows, err := r.db.Query(ctx, fmt.Sprintf(query, productColumns), ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []domain.Product
	for rows.Next() {
		var product domain.Product
		if err := rows.Scan(
			&product.Id,
			&product.CategoryId,
			&product.CategoryName,
			&product.CategoryActive,
			&product.Name,
			&product.Slug,
			&product.Description,
			&product.Price,
			&product.Stock,
			&product.StockReserved,
			&product.IsActive,
			&product.IsPromo,
			&product.CreatedAt,
			&product.UpdatedAt,
		); err != nil {
			return nil, err
		}
		products = append(products, product)
	}

	return products, rows.Err()
}

func (r *productRepository) ReserveStock(ctx context.Context, productId string, variantId string, quantity int32) error {
	if variantId != "" {
		query := `
			UPDATE product_variant
			SET stock_reserved = stock_reserved + $1
			WHERE id = $2 AND stock - stock_reserved >= $1 AND deleted_at IS NULL
		`
		result, err := r.db.Exec(ctx, query, quantity, variantId)
		if err != nil {
			return err
		}

		rowsAffected := result.RowsAffected()
		if rowsAffected == 0 {
			return errors.New(constants.ErrInsufficientStock)
		}

		return nil
	}

	query := `
		UPDATE product
		SET stock_reserved = stock_reserved + $1
		WHERE id = $2 AND stock - stock_reserved >= $1 AND deleted_at IS NULL
	`
	result, err := r.db.Exec(ctx, query, quantity, productId)
	if err != nil {
		return err
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.New(constants.ErrInsufficientStock)
	}

	return nil
}

func (r *productRepository) ReleaseStock(ctx context.Context, productId string, variantId string, quantity int32) error {
	if variantId != "" {
		query := `
			UPDATE product_variant
			SET stock_reserved = stock_reserved - $1
			WHERE id = $2 AND stock_reserved >= $1 AND deleted_at IS NULL
		`
		result, err := r.db.Exec(ctx, query, quantity, variantId)
		if err != nil {
			return err
		}

		if result.RowsAffected() == 0 {
			return errors.New(constants.ErrVariantNotFound)
		}

		return nil
	}

	query := `
		UPDATE product
		SET stock_reserved = stock_reserved - $1
		WHERE id = $2 AND stock_reserved >= $1 AND deleted_at IS NULL
	`
	result, err := r.db.Exec(ctx, query, quantity, productId)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return errors.New(constants.ErrProductNotFound)
	}

	return nil
}

func (r *productRepository) ConfirmStock(ctx context.Context, productId string, variantId string, quantity int32) error {
	if variantId != "" {
		query := `
			UPDATE product_variant
			SET stock = stock - $1, stock_reserved = GREATEST(0, stock_reserved - $1)
			WHERE id = $2 AND stock_reserved >= $1 AND deleted_at IS NULL
		`
		result, err := r.db.Exec(ctx, query, quantity, variantId)
		if err != nil {
			return err
		}

		if result.RowsAffected() == 0 {
			return errors.New(constants.ErrInsufficientStock)
		}

		return nil
	}

	query := `
		UPDATE product
		SET stock = stock - $1, stock_reserved = GREATEST(0, stock_reserved - $1)
		WHERE id = $2 AND stock_reserved >= $1 AND deleted_at IS NULL
	`
	result, err := r.db.Exec(ctx, query, quantity, productId)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return errors.New(constants.ErrInsufficientStock)
	}

	return nil
}
