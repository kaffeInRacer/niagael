package token

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"kaffein/auth-service/config"
	"kaffein/auth-service/utils/constants"
)

type Claims struct {
	SessionID uuid.UUID `json:"sid"`
	Roles     []string  `json:"roles"`
	TokenType string    `json:"token_type"`
	jwt.RegisteredClaims
}

type Manager struct{ cfg config.JWTConfig }

func NewManager(cfg config.JWTConfig) *Manager { return &Manager{cfg: cfg} }

func (m *Manager) Pair(userID uuid.UUID, role string) (string, string, uuid.UUID, time.Time, error) {
	sid := uuid.New()
	now := time.Now()
	accessExpiry := now.Add(m.cfg.AccessTTL)
	access, err := m.sign(userID, sid, role, "access", now, accessExpiry)
	if err != nil {
		return "", "", uuid.Nil, time.Time{}, err
	}
	refresh, err := m.sign(userID, sid, role, "refresh", now, now.Add(m.cfg.RefreshTTL))
	return access, refresh, sid, accessExpiry, err
}

func (m *Manager) sign(userID, sid uuid.UUID, role, tokenType string, issuedAt, expiresAt time.Time) (string, error) {
	claims := Claims{
		SessionID: sid, Roles: []string{role}, TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{Issuer: m.cfg.Issuer, Subject: userID.String(), IssuedAt: jwt.NewNumericDate(issuedAt), ExpiresAt: jwt.NewNumericDate(expiresAt)},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(m.cfg.Secret))
}

func (m *Manager) Parse(raw, tokenType string) (*Claims, error) {
	claims := &Claims{}
	tok, err := jwt.ParseWithClaims(raw, claims, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, errors.New(constants.ErrInvalidToken)
		}
		return []byte(m.cfg.Secret), nil
	}, jwt.WithIssuer(m.cfg.Issuer), jwt.WithExpirationRequired(), jwt.WithIssuedAt())
	if err != nil || !tok.Valid || claims.TokenType != tokenType || claims.SessionID == uuid.Nil || len(claims.Roles) == 0 {
		return nil, errors.New(constants.ErrInvalidToken)
	}
	if _, err := uuid.Parse(claims.Subject); err != nil {
		return nil, errors.New(constants.ErrInvalidToken)
	}
	return claims, nil
}
