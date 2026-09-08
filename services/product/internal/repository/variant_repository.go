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

type variantRepository struct {
	db postgresql.DBTX
}

func NewVariantRepository(store *postgresql.Store) IRepository.VariantRepository {
	return &variantRepository{
		db: store,
	}
}

func (r *variantRepository) ListByProductId(ctx context.Context, productId string) ([]domain.ProductVariant, error) {
	return r.listByProductIds(ctx, []string{productId})
}

func (r *variantRepository) ListByProductIds(ctx context.Context, productIds []string) ([]domain.ProductVariant, error) {
	return r.listByProductIds(ctx, productIds)
}

func (r *variantRepository) listByProductIds(ctx context.Context, productIds []string) ([]domain.ProductVariant, error) {
	const query = `
		SELECT id, product_id, name, price, stock, stock_reserved, attributes, is_active, created_at, updated_at
		FROM product_variant
		WHERE product_id = ANY($1) AND deleted_at IS NULL
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, query, productIds)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var variants []domain.ProductVariant
	for rows.Next() {
		var variant domain.ProductVariant
		if err := rows.Scan(
			&variant.Id,
			&variant.ProductId,
			&variant.Name,
			&variant.Price,
			&variant.Stock,
			&variant.StockReserved,
			&variant.Attributes,
			&variant.IsActive,
			&variant.CreatedAt,
			&variant.UpdatedAt,
		); err != nil {
			return nil, err
		}
		variants = append(variants, variant)
	}

	return variants, rows.Err()
}

func (r *variantRepository) Create(ctx context.Context, args dto.CreateVariantDto) error {
	const query = `
		INSERT INTO product_variant (id, product_id, name, price, stock, attributes, is_active, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
	`
	_, err := r.db.Exec(ctx, query, args.Id, args.ProductId, args.Name, args.Price, args.Stock, args.Attributes, args.IsActive)
	if err != nil {
		return err
	}

	return nil
}

func (r *variantRepository) Update(ctx context.Context, args dto.UpdateVariantDto) error {
	const query = `
		UPDATE product_variant
		SET name = $2, price = $3, stock = $4, attributes = $5, is_active = $6, updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`
	cmdTag, err := r.db.Exec(ctx, query, args.Id, args.Name, args.Price, args.Stock, args.Attributes, args.IsActive)
	if err != nil {
		return err
	}

	if cmdTag.RowsAffected() == 0 {
		return errors.New(constants.ErrVariantNotFound)
	}

	return nil
}

func (r *variantRepository) Delete(ctx context.Context, id string) error {
	const query = `
		UPDATE product_variant
		SET deleted_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`

	cmdTag, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	if cmdTag.RowsAffected() == 0 {
		return errors.New(constants.ErrVariantNotFound)
	}

	return nil
}

func (r *variantRepository) ReadById(ctx context.Context, id string) (*domain.ProductVariant, error) {
	const query = `
		SELECT id, product_id, name, price, stock, stock_reserved, attributes, is_active, created_at, updated_at
		FROM product_variant
		WHERE id = $1 AND deleted_at IS NULL
	`

	var variant domain.ProductVariant
	err := r.db.QueryRow(ctx, query, id).Scan(
		&variant.Id,
		&variant.ProductId,
		&variant.Name,
		&variant.Price,
		&variant.Stock,
		&variant.StockReserved,
		&variant.Attributes,
		&variant.IsActive,
		&variant.CreatedAt,
		&variant.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &variant, nil
}

func (r *variantRepository) ReserveStock(ctx context.Context, variantId string, quantity int32) error {
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

func (r *variantRepository) ReleaseStock(ctx context.Context, variantId string, quantity int32) error {
	const query = `
		UPDATE product_variant 
		SET stock_reserved = GREATEST(0, stock_reserved - $1) 
		WHERE id = $2 AND deleted_at IS NULL
	`
	_, err := r.db.Exec(ctx, query, quantity, variantId)
	return err
}

func (r *variantRepository) ConfirmStock(ctx context.Context, variantId string, quantity int32) error {
	const query = `
		UPDATE product_variant 
		SET stock = stock - $1, stock_reserved = GREATEST(0, stock_reserved - $1) 
		WHERE id = $2 AND stock_reserved >= $1 AND deleted_at IS NULL
	`
	_, err := r.db.Exec(ctx, query, quantity, variantId)
	return err
}
