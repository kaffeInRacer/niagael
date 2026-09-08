package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

const claimsKey = "auth.claims"

type Claims struct {
	SessionID string   `json:"sid"`
	Roles     []string `json:"roles"`
	TokenType string   `json:"token_type"`
	jwt.RegisteredClaims
}

type Service struct {
	secret   []byte
	issuer   string
	redis    *redis.Client
	enforcer *casbin.SyncedEnforcer
}

func New(ctx context.Context, secret, issuer string, redisDB int, db *pgxpool.Pool, redisClient *redis.Client) (*Service, error) {
	if secret == "" || issuer == "" {
		return nil, errors.New("JWT secret and issuer are required")
	}
	m := model.NewModel()
	m.AddDef("r", "r", "sub, obj, act")
	m.AddDef("p", "p", "sub, obj, act")
	m.AddDef("e", "e", "some(where (p_eft == allow))")
	m.AddDef("m", "m", "r.sub == p.sub && r.obj == p.obj && r.act == p.act")
	e, err := casbin.NewSyncedEnforcer(m)
	if err != nil {
		return nil, fmt.Errorf("create casbin enforcer: %w", err)
	}
	rows, err := db.Query(ctx, `SELECT v0, v1, v2 FROM casbin_rule WHERE ptype = 'p'`)
	if err != nil {
		return nil, fmt.Errorf("load casbin policies: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var role, resource, action string
		if err := rows.Scan(&role, &resource, &action); err != nil {
			return nil, fmt.Errorf("scan casbin policy: %w", err)
		}
		if _, err := e.AddPolicy(role, resource, action); err != nil {
			return nil, fmt.Errorf("add casbin policy: %w", err)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("load casbin policies: %w", err)
	}
	redisOptions := *redisClient.Options()
	redisOptions.DB = redisDB
	return &Service{secret: []byte(secret), issuer: issuer, redis: redis.NewClient(&redisOptions), enforcer: e}, nil
}

func (s *Service) Close() error { return s.redis.Close() }

func (s *Service) Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		rawToken, ok := accessToken(c.Request)
		if !ok {
			abort(c, http.StatusUnauthorized, "unauthorized")
			return
		}
		claims := &Claims{}
		token, err := jwt.ParseWithClaims(rawToken, claims, func(token *jwt.Token) (any, error) {
			if token.Method != jwt.SigningMethodHS256 {
				return nil, errors.New("unexpected signing method")
			}
			return s.secret, nil
		}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}), jwt.WithIssuer(s.issuer), jwt.WithExpirationRequired(), jwt.WithIssuedAt())
		if err != nil || !token.Valid || claims.Subject == "" || claims.SessionID == "" || claims.IssuedAt == nil || claims.TokenType != "access" || len(claims.Roles) == 0 {
			abort(c, http.StatusUnauthorized, "unauthorized")
			return
		}
		revoked, err := s.redis.Exists(c.Request.Context(), "auth:revoked:"+claims.SessionID).Result()
		if err != nil || revoked > 0 {
			abort(c, http.StatusUnauthorized, "unauthorized")
			return
		}
		c.Set(claimsKey, claims)
		c.Next()
	}
}

func accessToken(r *http.Request) (string, bool) {
	if cookie, err := r.Cookie("access_token"); err == nil {
		token := strings.TrimSpace(cookie.Value)
		return token, token != ""
	}

	parts := strings.SplitN(r.Header.Get("Authorization"), " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", false
	}
	token := strings.TrimSpace(parts[1])
	return token, token != ""
}

func (s *Service) Authorize(resource, action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, ok := ClaimsFrom(c)
		if !ok {
			abort(c, http.StatusUnauthorized, "unauthorized")
			return
		}
		for _, role := range claims.Roles {
			allowed, err := s.enforcer.Enforce(role, resource, action)
			if err == nil && allowed {
				c.Next()
				return
			}
		}
		abort(c, http.StatusForbidden, "forbidden")
	}
}

func (s *Service) RequireRoles(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		for _, role := range roles {
			if HasRole(c, role) {
				c.Next()
				return
			}
		}
		abort(c, http.StatusForbidden, "forbidden")
	}
}

func ClaimsFrom(c *gin.Context) (*Claims, bool) {
	value, ok := c.Get(claimsKey)
	if !ok {
		return nil, false
	}
	claims, ok := value.(*Claims)
	return claims, ok
}

func HasRole(c *gin.Context, wanted string) bool {
	claims, ok := ClaimsFrom(c)
	if !ok {
		return false
	}
	for _, role := range claims.Roles {
		if role == wanted {
			return true
		}
	}
	return false
}

func CanBypassOwnership(c *gin.Context) bool {
	return HasRole(c, "staff") || HasRole(c, "admin")
}

func RequireOwner(c *gin.Context, userID string) bool {
	claims, ok := ClaimsFrom(c)
	if !ok {
		abort(c, http.StatusUnauthorized, "unauthorized")
		return false
	}
	if CanBypassOwnership(c) || claims.Subject == userID {
		return true
	}
	abort(c, http.StatusForbidden, "forbidden")
	return false
}

func abort(c *gin.Context, status int, message string) {
	c.AbortWithStatusJSON(status, gin.H{"error": message})
}
