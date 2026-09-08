package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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

func Auth(tokens *token.Manager, sessions IRepository.SessionRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		raw, err := c.Cookie("access_token")
		if err != nil {
			parts := strings.Fields(c.GetHeader("Authorization"))
			if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
				raw = parts[1]
			}
		}
		if raw == "" {
			abort(c, http.StatusUnauthorized, constants.ErrUnauthorized)
			return
		}
		claims, err := tokens.Parse(raw, "access")
		if err != nil {
			abort(c, http.StatusUnauthorized, constants.ErrInvalidToken)
			return
		}
		revoked, err := sessions.IsRevoked(c.Request.Context(), claims.SessionID)
		if err != nil || revoked {
			abort(c, http.StatusUnauthorized, constants.ErrInvalidToken)
			return
		}
		userID, _ := uuid.Parse(claims.Subject)
		c.Set(UserIDKey, userID)
		c.Set(RolesKey, claims.Roles)
		c.Set(SessionIDKey, claims.SessionID)
		c.Set(ExpirationKey, claims.ExpiresAt.Unix())
		c.Next()
	}
}

func Admin() gin.HandlerFunc {
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

func abort(c *gin.Context, status int, message string) {
	c.AbortWithStatusJSON(status, gin.H{"error": message})
}
