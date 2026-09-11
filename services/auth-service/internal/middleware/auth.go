package middleware

import (
	"fmt"
	"context"
	"net/http"
	"strings"
	"sync"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
	"kaffein/auth-service/internal/interfaces/IRepository"
	"kaffein/auth-service/internal/token"
	"kaffein/auth-service/utils/constants"
)

const (
	UserIDKey     = "user_id"
	RolesKey      = "roles"
	SessionIDKey  = "sid"
	ExpirationKey = "exp"
)

type Service struct {
	tokens   *token.Manager
	sessions IRepository.SessionRepository
	users    IRepository.UserRepository
	db       *pgxpool.Pool
	logger   zerolog.Logger

	enforcer   *casbin.SyncedEnforcer
	enforcerMu sync.RWMutex
}

func NewService(tokens *token.Manager, sessions IRepository.SessionRepository, users IRepository.UserRepository, db *pgxpool.Pool, logger zerolog.Logger) *Service {
	s := &Service{
		tokens:   tokens,
		sessions: sessions,
		users:    users,
		db:       db,
		logger:   logger,
	}
	if err := s.ReloadPolicies(); err != nil {
		logger.Error().Err(err).Msg("failed to load casbin policies")
	}
	return s
}

func (s *Service) Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		raw := bearerToken(c.GetHeader("Authorization"))
		if raw == "" {
			raw, _ = c.Cookie("access_token")
			raw = strings.TrimSpace(raw)
		}
		if raw == "" {
			abort(c, http.StatusUnauthorized, constants.ErrUnauthorized)
			return
		}
		claims, err := s.tokens.Parse(raw, "access")
		if err != nil {
			abort(c, http.StatusUnauthorized, constants.ErrInvalidToken)
			return
		}
		revoked, err := s.sessions.IsRevoked(c.Request.Context(), claims.SessionID)
		if err != nil || revoked {
			abort(c, http.StatusUnauthorized, constants.ErrInvalidToken)
			return
		}
		userID, _ := uuid.Parse(claims.Subject)
		user, err := s.users.ByID(c.Request.Context(), userID)
		if err != nil || user == nil || !user.IsActive || len(claims.Roles) != 1 || claims.Roles[0] != user.Role {
			abort(c, http.StatusUnauthorized, constants.ErrInvalidToken)
			return
		}
		c.Set(UserIDKey, userID)
		c.Set(RolesKey, claims.Roles)
		c.Set(SessionIDKey, claims.SessionID)
		c.Set(ExpirationKey, claims.ExpiresAt.Unix())
		c.Next()
	}
}

func (s *Service) Admin() gin.HandlerFunc {
	return func(c *gin.Context) {
		roles, ok := c.Get(RolesKey)
		if !ok {
			abort(c, http.StatusForbidden, constants.ErrForbidden)
			return
		}
		for _, role := range roles.([]string) {
			if role == "admin" {
				c.Next()
				return
			}
		}
		abort(c, http.StatusForbidden, constants.ErrForbidden)
	}
}

func (s *Service) Authorize(resource, action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		roles, ok := c.Get(RolesKey)
		if !ok {
			abort(c, http.StatusForbidden, constants.ErrForbidden)
			return
		}

		s.enforcerMu.RLock()
		e := s.enforcer
		s.enforcerMu.RUnlock()

		for _, role := range roles.([]string) {
			if allowed, err := e.Enforce(role, resource, action); err == nil && allowed {
				c.Next()
				return
			}
		}
		abort(c, http.StatusForbidden, constants.ErrForbidden)
	}
}

func (s *Service) ReloadPolicies() error {
	m := model.NewModel()
	m.AddDef("r", "r", "sub, obj, act")
	m.AddDef("p", "p", "sub, obj, act")
	m.AddDef("e", "e", "some(where (p_eft == allow))")
	m.AddDef("m", "m", "r.sub == p.sub && r.obj == p.obj && r.act == p.act")

	e, err := casbin.NewSyncedEnforcer(m)
	if err != nil {
		return fmt.Errorf("create casbin enforcer: %w", err)
	}
	rows, err := s.db.Query(context.Background(), `SELECT v0, v1, v2 FROM casbin_rule WHERE ptype = 'p' AND service = 'auth'`)
	if err != nil {
		return fmt.Errorf("load casbin policies: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var role, resource, action string
		if err := rows.Scan(&role, &resource, &action); err != nil {
			return fmt.Errorf("scan casbin policy: %w", err)
		}
		if _, err := e.AddPolicy(role, resource, action); err != nil {
			return fmt.Errorf("add casbin policy: %w", err)
		}
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("load casbin policies: %w", err)
	}

	s.enforcerMu.Lock()
	s.enforcer = e
	s.enforcerMu.Unlock()
	return nil
}

func bearerToken(header string) string {
	parts := strings.Fields(header)
	if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
		return parts[1]
	}
	return ""
}

func abort(c *gin.Context, status int, message string) {
	c.AbortWithStatusJSON(status, gin.H{"error": message})
}
