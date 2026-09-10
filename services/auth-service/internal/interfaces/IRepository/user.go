package IRepository

import (
	"context"

	"github.com/google/uuid"
	"kaffein/auth-service/internal/domain"
)

type UserRepository interface {
	Create(context.Context, *domain.User) error
	ByEmail(context.Context, string) (*domain.User, error)
	ByID(context.Context, uuid.UUID) (*domain.User, error)
	List(context.Context, string, int, int) ([]domain.User, int64, error)
	UpdateRole(context.Context, uuid.UUID, string) (*domain.User, error)
	UpdateStatus(context.Context, uuid.UUID, bool) (*domain.User, error)
	UpdateEmail(context.Context, uuid.UUID, string) (*domain.User, error)
	UpdatePassword(context.Context, uuid.UUID, string) (*domain.User, error)
	Delete(context.Context, uuid.UUID) error
}
