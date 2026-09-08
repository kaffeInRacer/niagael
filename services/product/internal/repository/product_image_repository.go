package repository

import (
	"context"
	"kaffein/product-service/internal/domain"
	"kaffein/product-service/internal/dto"
	"kaffein/product-service/internal/interfaces/IRepository"
	"kaffein/product-service/pkg/postgresql"
)

type productImageRepository struct {
	db postgresql.DBTX
}

func NewProductImageRepository(store *postgresql.Store) IRepository.ProductImageRepository {
	return &productImageRepository{
		db: store,
	}
}

func (r *productImageRepository) Create(ctx context.Context, args dto.CreateProductImageDto) error {
	const query = `
		INSERT INTO product_image (id, product_id, file_name, sort_order, created_at)
		VALUES ($1, $2, $3, $4, NOW())
	`
	_, err := r.db.Exec(ctx, query, args.Id, args.ProductId, args.FileName, args.SortOrder)
	if err != nil {
		return err
	}

	return nil
}

func (r *productImageRepository) Delete(ctx context.Context, id string) error {
	const query = `
		UPDATE product_image
		SET deleted_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

func (r *productImageRepository) ListByProductId(ctx context.Context, productId string) ([]domain.ProductImage, error) {
	return r.listByProductIds(ctx, []string{productId})
}

func (r *productImageRepository) ListByProductIds(ctx context.Context, productIds []string) ([]domain.ProductImage, error) {
	return r.listByProductIds(ctx, productIds)
}

func (r *productImageRepository) listByProductIds(ctx context.Context, productIds []string) ([]domain.ProductImage, error) {
	const query = `
		SELECT id, product_id, file_name, sort_order
		FROM product_image
		WHERE product_id = ANY($1) AND deleted_at IS NULL
		ORDER BY sort_order ASC
	`
	rows, err := r.db.Query(ctx, query, productIds)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var images []domain.ProductImage
	for rows.Next() {
		var image domain.ProductImage
		if err := rows.Scan(
			&image.Id,
			&image.ProductId,
			&image.FileName,
			&image.SortOrder,
		); err != nil {
			return nil, err
		}
		images = append(images, image)
	}

	return images, rows.Err()
}
