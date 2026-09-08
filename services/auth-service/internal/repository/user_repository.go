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

type userRepository struct{ db *pgxpool.Pool }

func NewUserRepository(db *pgxpool.Pool) IRepository.UserRepository { return &userRepository{db: db} }

const userColumns = `id, email, password_hash, role, is_active, created_at, updated_at, deleted_at`

func scanUser(row pgx.Row) (*domain.User, error) {
	var user domain.User
	err := row.Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Role, &user.IsActive, &user.CreatedAt, &user.UpdatedAt, &user.DeletedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, errors.New(constants.ErrUserNotFound)
	}
	return &user, err
}

func (r *userRepository) Create(ctx context.Context, user *domain.User) error {
	err := r.db.QueryRow(ctx, `INSERT INTO users (id,email,password_hash,role) VALUES ($1,$2,$3,$4) RETURNING created_at,updated_at,is_active`, user.ID, user.Email, user.PasswordHash, user.Role).Scan(&user.CreatedAt, &user.UpdatedAt, &user.IsActive)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return errors.New(constants.ErrEmailExists)
	}
	return err
}

func (r *userRepository) ByEmail(ctx context.Context, email string) (*domain.User, error) {
	return scanUser(r.db.QueryRow(ctx, `SELECT `+userColumns+` FROM users WHERE email=$1 AND deleted_at IS NULL`, email))
}

func (r *userRepository) ByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	return scanUser(r.db.QueryRow(ctx, `SELECT `+userColumns+` FROM users WHERE id=$1 AND deleted_at IS NULL`, id))
}

func (r *userRepository) List(ctx context.Context, limit, offset int) ([]domain.User, int64, error) {
	var count int64
	if err := r.db.QueryRow(ctx, `SELECT count(*) FROM users WHERE deleted_at IS NULL`).Scan(&count); err != nil {
		return nil, 0, err
	}
	rows, err := r.db.Query(ctx, `SELECT `+userColumns+` FROM users WHERE deleted_at IS NULL ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	users := make([]domain.User, 0)
	for rows.Next() {
		user, err := scanUser(rows)
		if err != nil {
			return nil, 0, err
		}
		users = append(users, *user)
	}
	return users, count, rows.Err()
}

func (r *userRepository) UpdateRole(ctx context.Context, id uuid.UUID, role string) (*domain.User, error) {
	return scanUser(r.db.QueryRow(ctx, `UPDATE users SET role=$2,updated_at=now() WHERE id=$1 AND deleted_at IS NULL RETURNING `+userColumns, id, role))
}

func (r *userRepository) UpdateStatus(ctx context.Context, id uuid.UUID, active bool) (*domain.User, error) {
	return scanUser(r.db.QueryRow(ctx, `UPDATE users SET is_active=$2,updated_at=now() WHERE id=$1 AND deleted_at IS NULL RETURNING `+userColumns, id, active))
}
