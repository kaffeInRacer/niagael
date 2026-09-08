package repository

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"kaffein/product-service/internal/domain"
	"kaffein/product-service/internal/dto"
	"kaffein/product-service/internal/interfaces/IRepository"
	"kaffein/product-service/pkg/postgresql"
	"kaffein/product-service/utils/constants"
)

var ErrInsufficientStock = errors.New("insufficient stock")

type productRepository struct {
	db postgresql.DBTX
}

func NewProductRepository(store *postgresql.Store) IRepository.ProductRepository {
	return &productRepository{
		db: store,
	}
}

func (r *productRepository) List(ctx context.Context, params dto.ListProductParams) ([]domain.Product, error) {
	const query = `
		SELECT p.id, p.category_id, c.name as category_name, p.name, p.slug, 
		       p.description, p.price, p.stock, p.stock_reserved, 
		       p.is_active, p.is_promo_excluded, p.created_at, p.updated_at
		FROM product p
		LEFT JOIN category c ON p.category_id = c.id
		WHERE (NULLIF($1::text, '') IS NULL OR p.name ILIKE '%' || $1::text || '%')
		AND ($2::boolean IS NULL OR p.is_active = $2::boolean)
		AND ($3::boolean IS NULL OR p.is_promo_excluded = $3::boolean)
		AND (NULLIF($4::text, '') IS NULL OR p.category_id = NULLIF($4::text, '')::uuid)
		AND (NULLIF($5::text, '') IS NULL OR p.price >= NULLIF($5::text, '')::numeric)
		AND (NULLIF($6::text, '') IS NULL OR p.price <= NULLIF($6::text, '')::numeric)
		AND p.deleted_at IS NULL
		ORDER BY
		  CASE WHEN $7::text = 'name'       AND $8::text = 'asc'  THEN p.name       END ASC,
		  CASE WHEN $7::text = 'price'      AND $8::text = 'asc'  THEN p.price      END ASC,
		  CASE WHEN $7::text = 'created_at' AND $8::text = 'asc'  THEN p.created_at END ASC,
		  CASE WHEN $7::text = 'updated_at' AND $8::text = 'asc'  THEN p.updated_at END ASC,
		  CASE WHEN $7::text = 'name'       AND $8::text = 'desc' THEN p.name       END DESC,
		  CASE WHEN $7::text = 'price'      AND $8::text = 'desc' THEN p.price      END DESC,
		  CASE WHEN $7::text = 'created_at' AND $8::text = 'desc' THEN p.created_at END DESC,
		  CASE WHEN $7::text = 'updated_at' AND $8::text = 'desc' THEN p.updated_at END DESC,
		p.created_at DESC
		LIMIT $9::int
		OFFSET $10::int
	`
	rows, err := r.db.Query(ctx, query,
		params.Search,
		params.IsActive,
		params.IsPromoExcluded,
		params.CategoryID,
		params.MinPrice,
		params.MaxPrice,
		params.OrderBy,
		params.OrderDir,
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
		FROM product p
		WHERE (NULLIF($1::text, '') IS NULL OR p.name ILIKE '%' || $1::text || '%')
		AND ($2::boolean IS NULL OR p.is_active = $2::boolean)
		AND ($3::boolean IS NULL OR p.is_promo_excluded = $3::boolean)
		AND (NULLIF($4::text, '') IS NULL OR p.category_id = NULLIF($4::text, '')::uuid)
		AND (NULLIF($5::text, '') IS NULL OR p.price >= NULLIF($5::text, '')::numeric)
		AND (NULLIF($6::text, '') IS NULL OR p.price <= NULLIF($6::text, '')::numeric)
		AND p.deleted_at IS NULL
	`

	var count int64
	err := r.db.QueryRow(ctx, query,
		params.Search,
		params.IsActive,
		params.IsPromoExcluded,
		params.CategoryID,
		params.MinPrice,
		params.MaxPrice,
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
		SELECT p.id, p.category_id, c.name as category_name, p.name, p.slug, 
		       p.description, p.price, p.stock, p.stock_reserved, 
		       p.is_active, p.is_promo_excluded, p.created_at, p.updated_at
		FROM product p
		LEFT JOIN category c ON p.category_id = c.id
		WHERE p.id = $1 AND p.deleted_at IS NULL
	`

	var product domain.Product
	err := r.db.QueryRow(ctx, query, id).Scan(
		&product.Id,
		&product.CategoryId,
		&product.CategoryName,
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
		SELECT p.id, p.category_id, c.name as category_name, p.name, p.slug, 
		       p.description, p.price, p.stock, p.stock_reserved, 
		       p.is_active, p.is_promo_excluded, p.created_at, p.updated_at
		FROM product p
		LEFT JOIN category c ON p.category_id = c.id
		WHERE p.slug = $1 AND p.deleted_at IS NULL
	`

	var product domain.Product
	err := r.db.QueryRow(ctx, query, slug).Scan(
		&product.Id,
		&product.CategoryId,
		&product.CategoryName,
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
		SELECT p.id, p.category_id, c.name as category_name, p.name, p.slug, 
		       p.description, p.price, p.stock, p.stock_reserved, 
		       p.is_active, p.is_promo_excluded, p.created_at, p.updated_at
		FROM product p
		LEFT JOIN category c ON p.category_id = c.id
		WHERE p.id = ANY($1) AND p.is_active = true AND p.deleted_at IS NULL
	`
	rows, err := r.db.Query(ctx, query, ids)
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
			return ErrInsufficientStock
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
		return ErrInsufficientStock
	}
	return nil
}

func (r *productRepository) ReleaseStock(ctx context.Context, productId string, variantId string, quantity int32) error {
	if variantId != "" {
		query := `
			UPDATE product_variant 
			SET stock_reserved = GREATEST(0, stock_reserved - $1) 
			WHERE id = $2 AND deleted_at IS NULL
		`
		_, err := r.db.Exec(ctx, query, quantity, variantId)
		return err
	}

	query := `
		UPDATE product 
		SET stock_reserved = GREATEST(0, stock_reserved - $1) 
		WHERE id = $2 AND deleted_at IS NULL
	`
	_, err := r.db.Exec(ctx, query, quantity, productId)
	return err
}

func (r *productRepository) ConfirmStock(ctx context.Context, productId string, variantId string, quantity int32) error {
	if variantId != "" {
		query := `
			UPDATE product_variant 
			SET stock = stock - $1, stock_reserved = GREATEST(0, stock_reserved - $1) 
			WHERE id = $2 AND stock_reserved >= $1 AND deleted_at IS NULL
		`
		_, err := r.db.Exec(ctx, query, quantity, variantId)
		return err
	}

	query := `
		UPDATE product 
		SET stock = stock - $1, stock_reserved = GREATEST(0, stock_reserved - $1) 
		WHERE id = $2 AND stock_reserved >= $1 AND deleted_at IS NULL
	`
	_, err := r.db.Exec(ctx, query, quantity, productId)
	return err
}
