package usecase

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"kaffein/auth-service/config"
	"kaffein/auth-service/internal/domain"
	"kaffein/auth-service/internal/dto"
	"kaffein/auth-service/internal/interfaces/IRepository"
	"kaffein/auth-service/internal/interfaces/IUseCase"
	"kaffein/auth-service/internal/token"
	"kaffein/auth-service/utils/constants"
)

type authUseCase struct {
	users    IRepository.UserRepository
	sessions IRepository.SessionRepository
	tokens   *token.Manager
	cfg      config.JWTConfig
}

func NewAuthUseCase(users IRepository.UserRepository, sessions IRepository.SessionRepository, tokens *token.Manager, cfg config.JWTConfig) IUseCase.AuthUseCase {
	return &authUseCase{users: users, sessions: sessions, tokens: tokens, cfg: cfg}
}

func (u *authUseCase) Register(ctx context.Context, req dto.RegisterRequest) (*domain.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	user := &domain.User{ID: uuid.New(), Email: strings.ToLower(strings.TrimSpace(req.Email)), PasswordHash: string(hash), Role: domain.RoleTenant}
	if err := u.users.Create(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

func (u *authUseCase) Login(ctx context.Context, req dto.LoginRequest) (*dto.TokenResponse, error) {
	user, err := u.users.ByEmail(ctx, strings.TrimSpace(req.Email))
	if err != nil || bcrypt.CompareHashAndPassword([]byte(valueOrEmpty(user)), []byte(req.Password)) != nil {
		return nil, errors.New(constants.ErrInvalidCredentials)
	}
	if !user.IsActive {
		return nil, errors.New(constants.ErrUserDisabled)
	}
	return u.issue(ctx, user)
}

func valueOrEmpty(user *domain.User) string {
	if user == nil {
		return ""
	}
	return user.PasswordHash
}

func (u *authUseCase) issue(ctx context.Context, user *domain.User) (*dto.TokenResponse, error) {
	access, refresh, sid, accessExpiry, err := u.tokens.Pair(user.ID, user.Role)
	if err != nil {
		return nil, err
	}
	session := IRepository.Session{UserID: user.ID, Role: user.Role, RefreshToken: refresh, AccessExpiresAt: accessExpiry}
	if err := u.sessions.Save(ctx, sid, session, u.cfg.RefreshTTL); err != nil {
		return nil, err
	}
	return &dto.TokenResponse{AccessToken: access, RefreshToken: refresh, TokenType: "Bearer", ExpiresIn: int64(u.cfg.AccessTTL.Seconds())}, nil
}

func (u *authUseCase) Refresh(ctx context.Context, raw string) (*dto.TokenResponse, error) {
	claims, err := u.tokens.Parse(raw, "refresh")
	if err != nil {
		return nil, err
	}
	session, err := u.sessions.Get(ctx, claims.SessionID)
	if err != nil || session.RefreshToken != raw {
		return nil, errors.New(constants.ErrInvalidToken)
	}
	user, err := u.users.ByID(ctx, session.UserID)
	if err != nil {
		return nil, err
	}
	if !user.IsActive {
		return nil, errors.New(constants.ErrUserDisabled)
	}
	if err := u.sessions.Revoke(ctx, claims.SessionID); err != nil {
		return nil, err
	}
	return u.issue(ctx, user)
}

func (u *authUseCase) Logout(ctx context.Context, sid uuid.UUID, expiration int64) error {
	ttl := time.Until(time.Unix(expiration, 0))
	return u.sessions.RevokeWithTTL(ctx, sid, ttl)
}

func (u *authUseCase) Me(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	return u.users.ByID(ctx, id)
}

type adminUseCase struct {
	users    IRepository.UserRepository
	sessions IRepository.SessionRepository
}

func NewAdminUseCase(users IRepository.UserRepository, sessions IRepository.SessionRepository) IUseCase.AdminUseCase {
	return &adminUseCase{users: users, sessions: sessions}
}

func (u *adminUseCase) ListUsers(ctx context.Context, page, size int) (*dto.UserList, error) {
	users, count, err := u.users.List(ctx, size, (page-1)*size)
	return &dto.UserList{Data: users, Count: count, Page: page, PageSize: size}, err
}

func (u *adminUseCase) ChangeRole(ctx context.Context, id uuid.UUID, role string) (*domain.User, error) {
	if !domain.ValidRole(role) {
		return nil, errors.New(constants.ErrInvalidRole)
	}
	user, err := u.users.UpdateRole(ctx, id, role)
	if err != nil {
		return nil, err
	}
	if err := u.sessions.RevokeUser(ctx, id); err != nil {
		return nil, err
	}
	return user, nil
}

func (u *adminUseCase) ChangeStatus(ctx context.Context, id uuid.UUID, active bool) (*domain.User, error) {
	user, err := u.users.UpdateStatus(ctx, id, active)
	if err != nil {
		return nil, err
	}
	if err := u.sessions.RevokeUser(ctx, id); err != nil {
		return nil, err
	}
	return user, nil
}
