package IUseCase

import (
	"context"

	"github.com/google/uuid"
	"kaffein/auth-service/internal/domain"
	"kaffein/auth-service/internal/dto"
)

type AuthUseCase interface {
	Register(context.Context, dto.RegisterRequest) (*domain.User, error)
	Login(context.Context, dto.LoginRequest) (*dto.TokenResponse, error)
	Refresh(context.Context, string) (*dto.TokenResponse, error)
	Logout(context.Context, uuid.UUID, int64) error
	Me(context.Context, uuid.UUID) (*domain.User, error)
}

type AdminUseCase interface {
	ListUsers(context.Context, int, int) (*dto.UserList, error)
	ChangeRole(context.Context, uuid.UUID, string) (*domain.User, error)
	ChangeStatus(context.Context, uuid.UUID, bool) (*domain.User, error)
}
