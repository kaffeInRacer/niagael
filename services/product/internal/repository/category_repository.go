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

type categoryRepository struct {
	db postgresql.DBTX
}

func NewCategoryRepository(store *postgresql.Store) IRepository.CategoryRepository {
	return &categoryRepository{
		db: store,
	}
}

// categoryOrderBy builds a safe ORDER BY clause for category queries.
func categoryOrderBy(column, direction string, prefix string) (string, string) {
	allowed := map[string]string{
		"name":       prefix + ".name",
		"slug":       prefix + ".slug",
		"created_at": prefix + ".created_at",
		"updated_at": prefix + ".updated_at",
	}

	col, ok := allowed[column]
	if !ok {
		col = prefix + ".created_at"
	}

	dir := "DESC"
	if strings.EqualFold(direction, "asc") {
		dir = "ASC"
	}

	orderBy := fmt.Sprintf("%s %s", col, dir)
	fallback := prefix + ".created_at DESC"
	return orderBy, fallback
}

func (r *categoryRepository) List(ctx context.Context, params dto.ListCategoryParams) ([]domain.Category, error) {
	orderBy, fallbackOrderBy := categoryOrderBy(params.OrderBy, params.OrderDir, "c")

	query := fmt.Sprintf(`
		SELECT id, name, slug, description, is_active, created_at, updated_at
		FROM category
		WHERE (NULLIF($1::text, '') IS NULL OR name ILIKE '%%' || $1::text || '%%')
		AND ($2::boolean IS NULL OR is_active = $2::boolean)
		AND deleted_at IS NULL
		ORDER BY %s, %s
		LIMIT $3::int
		OFFSET $4::int
	`, orderBy, fallbackOrderBy)

	rows, err := r.db.Query(ctx, query,
		params.Search,
		params.IsActive,
		params.PageSize,
		params.PageOffset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []domain.Category
	for rows.Next() {
		var category domain.Category
		if err := rows.Scan(
			&category.Id,
			&category.Name,
			&category.Slug,
			&category.Description,
			&category.IsActive,
			&category.CreatedAt,
			&category.UpdatedAt,
		); err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}

	return categories, rows.Err()
}

func (r *categoryRepository) ListCount(ctx context.Context, params dto.ListCategoryParams) (int64, error) {
	const query = `
		SELECT COUNT(*)
		FROM category
		WHERE (NULLIF($1::text, '') IS NULL OR name ILIKE '%' || $1::text || '%')
		AND ($2::boolean IS NULL OR is_active = $2::boolean)
		AND deleted_at IS NULL
	`

	var count int64
	err := r.db.QueryRow(ctx, query,
		params.Search,
		params.IsActive,
	).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (r *categoryRepository) Create(ctx context.Context, args dto.CreateCategoryDto) error {
	const query = `
		INSERT INTO category (id, name, slug, description, created_at)
		VALUES ($1, $2, $3, $4, NOW())
	`
	_, err := r.db.Exec(ctx, query, args.Id, args.Name, args.Slug, args.Description)
	if err != nil {
		return err
	}

	return nil
}

func (r *categoryRepository) Update(ctx context.Context, args dto.UpdateCategoryDto) error {
	const query = `
		UPDATE category
		SET name = $2, slug = $3, description = $4, updated_at = NOW()
		WHERE id = $1
	`
	cmdTag, err := r.db.Exec(ctx, query, args.Id, args.Name, args.Slug, args.Description)
	if err != nil {
		return err
	}

	if cmdTag.RowsAffected() == 0 {
		return errors.New(constants.ErrCategoryNotFound)
	}

	return nil
}

func (r *categoryRepository) Delete(ctx context.Context, id string) error {
	const query = `
		UPDATE category
		SET deleted_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`

	cmdTag, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	if cmdTag.RowsAffected() == 0 {
		return errors.New(constants.ErrCategoryNotFound)
	}

	return nil
}

func (r *categoryRepository) ReadById(ctx context.Context, id string) (*domain.Category, error) {
	const query = `
		SELECT id, name, slug, description, is_active, created_at, updated_at
		FROM category
		WHERE id = $1 AND deleted_at IS NULL
	`

	var category domain.Category
	err := r.db.QueryRow(ctx, query, id).Scan(
		&category.Id,
		&category.Name,
		&category.Slug,
		&category.Description,
		&category.IsActive,
		&category.CreatedAt,
		&category.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &category, nil
}

func (r *categoryRepository) ReadBySlug(ctx context.Context, slug string) (*domain.Category, error) {
	const query = `
		SELECT id, name, slug, description, is_active, created_at, updated_at
		FROM category
		WHERE slug = $1 AND deleted_at IS NULL
	`

	var category domain.Category
	err := r.db.QueryRow(ctx, query, slug).Scan(
		&category.Id,
		&category.Name,
		&category.Slug,
		&category.Description,
		&category.IsActive,
		&category.CreatedAt,
		&category.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &category, nil
}

func (r *categoryRepository) ListWithProductCount(ctx context.Context, params dto.ListCategoryParams) ([]domain.Category, error) {
	orderBy, fallbackOrderBy := categoryOrderBy(params.OrderBy, params.OrderDir, "c")

	query := fmt.Sprintf(`
		SELECT 
			c.id, c.name, c.slug, c.description, c.is_active, c.created_at, c.updated_at,
			COUNT(p.id) as product_count
		FROM category c
		LEFT JOIN product p ON p.category_id = c.id AND p.deleted_at IS NULL AND (NOT $5::boolean OR p.is_active = true)
		WHERE (NULLIF($1::text, '') IS NULL OR c.name ILIKE '%%' || $1::text || '%%')
		AND ($2::boolean IS NULL OR c.is_active = $2::boolean)
		AND c.deleted_at IS NULL
		GROUP BY c.id, c.name, c.slug, c.description, c.is_active, c.created_at, c.updated_at
		ORDER BY %s, %s
		LIMIT $3::int
		OFFSET $4::int
	`, orderBy, fallbackOrderBy)

	rows, err := r.db.Query(ctx, query,
		params.Search,
		params.IsActive,
		params.PageSize,
		params.PageOffset,
		params.ActiveProductsOnly,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []domain.Category
	for rows.Next() {
		var category domain.Category
		var productCount int64
		if err := rows.Scan(
			&category.Id,
			&category.Name,
			&category.Slug,
			&category.Description,
			&category.IsActive,
			&category.CreatedAt,
			&category.UpdatedAt,
			&productCount,
		); err != nil {
			return nil, err
		}
		category.ProductCount = &productCount
		categories = append(categories, category)
	}

	return categories, rows.Err()
}
