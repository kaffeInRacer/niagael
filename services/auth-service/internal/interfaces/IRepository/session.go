package IRepository

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Session struct {
	UserID          uuid.UUID `json:"user_id"`
	Role            string    `json:"role"`
	RefreshToken    string    `json:"refresh_token"`
	AccessExpiresAt time.Time `json:"access_expires_at"`
}

type SessionRepository interface {
	Save(context.Context, uuid.UUID, Session, time.Duration) error
	Get(context.Context, uuid.UUID) (*Session, error)
	Revoke(context.Context, uuid.UUID) error
	RevokeWithTTL(context.Context, uuid.UUID, time.Duration) error
	IsRevoked(context.Context, uuid.UUID) (bool, error)
	RevokeUser(context.Context, uuid.UUID) error
}
