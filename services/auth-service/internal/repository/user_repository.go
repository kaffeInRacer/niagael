package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"kaffein/auth-service/internal/domain"
	"kaffein/auth-service/internal/interfaces/IRepository"
	"kaffein/auth-service/utils/constants"
)

type userRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) IRepository.UserRepository {
	return &userRepository{
		db: db,
	}
}

func (r *userRepository) Create(ctx context.Context, user *domain.User) error {
	const query = `
		INSERT INTO users (id, email, password_hash, role)
		VALUES ($1, $2, $3, $4)
		RETURNING created_at, updated_at, is_active
	`

	err := r.db.QueryRow(ctx, query, user.ID, user.Email, user.PasswordHash, user.Role).Scan(
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.IsActive)

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return errors.New(constants.ErrEmailExists)
	}

	return err
}

func (r *userRepository) ByEmail(ctx context.Context, email string) (*domain.User, error) {
	const query = `
		SELECT id, email, password_hash, role, is_active, created_at, updated_at, deleted_at
		FROM users
		WHERE email = $1 AND deleted_at IS NULL
	`

	var user domain.User
	err := r.db.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.DeletedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, errors.New(constants.ErrUserNotFound)
	}

	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *userRepository) ByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	const query = `
		SELECT id, email, password_hash, role, is_active, created_at, updated_at, deleted_at
		FROM users
		WHERE id = $1 AND deleted_at IS NULL
	`

	var user domain.User
	err := r.db.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.DeletedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, errors.New(constants.ErrUserNotFound)
	}

	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *userRepository) List(ctx context.Context, limit, offset int) ([]domain.User, int64, error) {
	const countQuery = `SELECT count(*) FROM users WHERE deleted_at IS NULL`

	var count int64
	if err := r.db.QueryRow(ctx, countQuery).Scan(&count); err != nil {
		return nil, 0, err
	}

	const query = `
		SELECT id, email, password_hash, role, is_active, created_at, updated_at, deleted_at
		FROM users
		WHERE deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`
	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	users := make([]domain.User, 0)
	for rows.Next() {
		var user domain.User
		if err := rows.Scan(
			&user.ID,
			&user.Email,
			&user.PasswordHash,
			&user.Role,
			&user.IsActive,
			&user.CreatedAt,
			&user.UpdatedAt,
			&user.DeletedAt,
		); err != nil {
			return nil, 0, err
		}
		users = append(users, user)
}

	return users, count, rows.Err()
}

func (r *userRepository) UpdateRole(ctx context.Context, id uuid.UUID, role string) (*domain.User, error) {
	const query = `
		UPDATE users
		SET role = $2, updated_at = now()
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING id, email, password_hash, role, is_active, created_at, updated_at, deleted_at
	`

	var user domain.User
	err := r.db.QueryRow(ctx, query, id, role).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.DeletedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, errors.New(constants.ErrUserNotFound)
	}

	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *userRepository) UpdateStatus(ctx context.Context, id uuid.UUID, active bool) (*domain.User, error) {
	const query = `
		UPDATE users
		SET is_active = $2, updated_at = now()
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING id, email, password_hash, role, is_active, created_at, updated_at, deleted_at
	`

	var user domain.User
	err := r.db.QueryRow(ctx, query, id, active).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.DeletedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, errors.New(constants.ErrUserNotFound)
	}

	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *userRepository) UpdateEmail(ctx context.Context, id uuid.UUID, email string) (*domain.User, error) {
	const query = `
		UPDATE users
		SET email = $2, updated_at = now()
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING id, email, password_hash, role, is_active, created_at, updated_at, deleted_at
	`

	var user domain.User
	err := r.db.QueryRow(ctx, query, id, email).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.DeletedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, errors.New(constants.ErrUserNotFound)
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return nil, errors.New(constants.ErrEmailExists)
	}

	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *userRepository) Delete(ctx context.Context, id uuid.UUID) error {
	const query = `
		UPDATE users
		SET deleted_at = now(), updated_at = now()
		WHERE id = $1 AND deleted_at IS NULL
	`

	tag, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errors.New(constants.ErrUserNotFound)
	}

	return nil
}
