package middleware

import (
	"net/http"
	"strings"
	"sync"

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

	mu     sync.Mutex
	policy map[string]map[string]map[string]bool
}

func NewService(tokens *token.Manager, sessions IRepository.SessionRepository, users IRepository.UserRepository, db *pgxpool.Pool, logger zerolog.Logger) *Service {
	return &Service{
		tokens:   tokens,
		sessions: sessions,
		users:    users,
		db:       db,
		logger:   logger,
	}
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

		var allowed bool
		err := s.db.QueryRow(c.Request.Context(), `SELECT EXISTS (
			SELECT 1 FROM casbin_rule
			WHERE ptype = 'p' AND service = 'auth' AND v0 = ANY($1) AND v1 = $2 AND v2 = $3
		)`, roles.([]string), resource, action).Scan(&allowed)
		if err != nil {
			s.logger.Error().Err(err).Str("resource", resource).Str("action", action).Msg("RBAC authorization query failed")
			abort(c, http.StatusInternalServerError, constants.ErrInternalServer)
			return
		}
		if !allowed {
			abort(c, http.StatusForbidden, constants.ErrForbidden)
			return
		}
		c.Next()
	}
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
