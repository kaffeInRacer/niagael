package auth

import (
	"context"
	"errors"
	"fmt"
	"kaffein/order-service/config"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"

	"kaffein/order-service/pkg/kafkawatcher"
	"kaffein/order-service/utils/constants"
)

const claimsKey = "auth.claims"

const (
	accessTokenType  = "access"
	roleStaff        = "staff"
	roleAdmin        = "admin"
	serviceName      = "order"
	policyNotifyChan = "casbin_policy_changed"
)

type Claims struct {
	SessionID string   `json:"sid"`
	Roles     []string `json:"roles"`
	TokenType string   `json:"token_type"`
	jwt.RegisteredClaims
}

func (c *Claims) isValid() bool {
	return c.Subject != "" &&
		c.SessionID != "" &&
		c.IssuedAt != nil &&
		c.TokenType == accessTokenType &&
		len(c.Roles) > 0
}

func (c *Claims) hasRole(wanted string) bool {
	for _, role := range c.Roles {
		if role == wanted {
			return true
		}
	}
	return false
}

type Service struct {
	secret       []byte
	issuer       string
	redis        *redis.Client
	db           *pgxpool.Pool
	enforcer     *casbin.SyncedEnforcer
	enforcerMu   sync.RWMutex
	reloadMu     sync.Mutex
	logger       zerolog.Logger
	cancelReload context.CancelFunc
	reloadDone   chan struct{}
	watcher      *kafkawatcher.Watcher
}

func New(ctx context.Context, c *config.Config, db *pgxpool.Pool, redisClient *redis.Client, logger zerolog.Logger) (*Service, error) {
	if c.JWT.Secret == "" || c.JWT.Issuer == "" {
		return nil, errors.New("JWT secret and issuer are required")
	}

	enforcer, err := loadEnforcer(ctx, db)
	if err != nil {
		return nil, err
	}

	redisOptions := *redisClient.Options()
	redisOptions.DB = c.JWT.RedisDB

	s := &Service{
		secret:   []byte(c.JWT.Secret),
		issuer:   c.JWT.Issuer,
		redis:    redis.NewClient(&redisOptions),
		db:       db,
		enforcer: enforcer,
		logger:   logger,
	}
	if c.RBAC.PolicyReloadInterval <= 0 {
		c.RBAC.PolicyReloadInterval = 30 * time.Second
	}

	reloadCtx, cancel := context.WithCancel(context.Background())
	s.cancelReload = cancel
	s.reloadDone = make(chan struct{})

	go s.watchPolicyChanges(reloadCtx, db)
	go s.autoReload(reloadCtx, db, c.RBAC.PolicyReloadInterval)
	if len(c.Kafka.Brokers) > 0 && c.Kafka.CasbinTopic != "" {
		s.watcher = kafkawatcher.New(c.Kafka.Brokers, c.Kafka.CasbinTopic, c.Kafka.GroupID)
		if err := s.watcher.SetUpdateCallback(func(string) { _ = s.reloadPolicies(context.Background(), db) }); err != nil {
			return nil, fmt.Errorf("start casbin watcher: %w", err)
		}
	}

	return s, nil
}

func loadEnforcer(ctx context.Context, db *pgxpool.Pool) (*casbin.SyncedEnforcer, error) {
	m := model.NewModel()
	m.AddDef("r", "r", "sub, obj, act")
	m.AddDef("p", "p", "sub, obj, act")
	m.AddDef("e", "e", "some(where (p_eft == allow))")
	m.AddDef("m", "m", "r.sub == p.sub && r.obj == p.obj && r.act == p.act")

	e, err := casbin.NewSyncedEnforcer(m)
	if err != nil {
		return nil, fmt.Errorf("create casbin enforcer: %w", err)
	}
	rows, err := db.Query(ctx, `SELECT v0, v1, v2 FROM casbin_rule WHERE ptype = 'p' AND service = $1`, serviceName)
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

	return e, nil
}

func (s *Service) reloadPolicies(ctx context.Context, db *pgxpool.Pool) error {
	s.reloadMu.Lock()
	defer s.reloadMu.Unlock()
	e, err := loadEnforcer(ctx, db)
	if err != nil {
		return err
	}
	s.enforcerMu.Lock()
	s.enforcer = e
	s.enforcerMu.Unlock()
	return nil
}

func (s *Service) watchPolicyChanges(ctx context.Context, db *pgxpool.Pool) {
	defer close(s.reloadDone)

	for {
		if ctx.Err() != nil {
			return
		}
		if err := s.listenLoop(ctx, db); err != nil && ctx.Err() == nil {
			s.logger.Error().Err(err).Str("service", serviceName).Msg("RBAC policy notify listener disconnected; reconnecting")
		}
		if ctx.Err() != nil {
			return
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(2 * time.Second):
		}
	}
}

func (s *Service) listenLoop(ctx context.Context, db *pgxpool.Pool) error {
	conn, err := db.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("acquire notify connection: %w", err)
	}
	defer conn.Release()

	if _, err := conn.Exec(ctx, "LISTEN "+policyNotifyChan); err != nil {
		return fmt.Errorf("listen policy channel: %w", err)
	}

	for {
		notification, err := conn.Conn().WaitForNotification(ctx)
		if err != nil {
			return fmt.Errorf("wait for policy notification: %w", err)
		}
		if notification != nil && notification.Channel == policyNotifyChan {
			if err := s.reloadPolicies(context.WithoutCancel(ctx), db); err != nil {
				s.logger.Error().Err(err).Str("service", serviceName).Msg("RBAC policy reload failed after notification")
			}
		}
	}
}

func (s *Service) autoReload(ctx context.Context, db *pgxpool.Pool, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := s.reloadPolicies(ctx, db); err != nil && ctx.Err() == nil {
				s.logger.Error().Err(err).Str("service", serviceName).Msg("RBAC policy reload failed; retrying")
			}
		}
	}
}

func (s *Service) Close() error {
	if s.cancelReload != nil {
		s.cancelReload()
	}
	if s.watcher != nil {
		s.watcher.Close()
	}
	select {
	case <-s.reloadDone:
	case <-time.After(5 * time.Second):
		s.logger.Error().Str("service", serviceName).Msg("timed out stopping RBAC policy poller")
	}
	return s.redis.Close()
}

func (s *Service) Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		rawToken, ok := accessToken(c.Request)
		if !ok {
			abort(c, http.StatusUnauthorized, constants.ErrUnauthorized)
			return
		}

		claims := &Claims{}
		token, err := jwt.ParseWithClaims(rawToken, claims, func(token *jwt.Token) (any, error) {
			if token.Method != jwt.SigningMethodHS256 {
				return nil, errors.New("unexpected signing method")
			}
			return s.secret, nil
		}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}), jwt.WithIssuer(s.issuer), jwt.WithExpirationRequired(), jwt.WithIssuedAt())

		if err != nil || !token.Valid || !claims.isValid() {
			abort(c, http.StatusUnauthorized, constants.ErrUnauthorized)
			return
		}

		revoked, err := s.redis.Exists(c.Request.Context(), "auth:revoked:"+claims.SessionID).Result()
		if err != nil || revoked > 0 {
			abort(c, http.StatusUnauthorized, constants.ErrUnauthorized)
			return
		}
		var role string
		var active bool
		if err := s.db.QueryRow(c.Request.Context(), `SELECT role, is_active FROM users WHERE id = $1 AND deleted_at IS NULL`, claims.Subject).Scan(&role, &active); err != nil || !active || len(claims.Roles) != 1 || claims.Roles[0] != role {
			abort(c, http.StatusUnauthorized, constants.ErrUnauthorized)
			return
		}

		c.Set(claimsKey, claims)
		c.Next()
	}
}

func accessToken(r *http.Request) (string, bool) {
	parts := strings.SplitN(r.Header.Get("Authorization"), " ", 2)
	if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
		token := strings.TrimSpace(parts[1])
		if token != "" {
			return token, true
		}
	}
	if cookie, err := r.Cookie("access_token"); err == nil {
		token := strings.TrimSpace(cookie.Value)
		return token, token != ""
	}
	return "", false
}

func (s *Service) Authorize(resource, action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, ok := ClaimsFrom(c)
		if !ok {
			abort(c, http.StatusUnauthorized, constants.ErrUnauthorized)
			return
		}

		s.enforcerMu.RLock()
		e := s.enforcer
		s.enforcerMu.RUnlock()

		for _, role := range claims.Roles {
			if allowed, err := e.Enforce(role, resource, action); err == nil && allowed {
				c.Next()
				return
			}
		}
		abort(c, http.StatusForbidden, constants.ErrForbidden)
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
		abort(c, http.StatusForbidden, constants.ErrForbidden)
	}
}

func ClaimsFrom(c *gin.Context) (*Claims, bool) {
	claims, ok := c.Get(claimsKey)
	if !ok {
		return nil, false
	}
	cl, ok := claims.(*Claims)
	return cl, ok
}

func HasRole(c *gin.Context, wanted string) bool {
	claims, ok := ClaimsFrom(c)
	return ok && claims.hasRole(wanted)
}

func CanBypassOwnership(c *gin.Context) bool {
	return HasRole(c, roleStaff) || HasRole(c, roleAdmin)
}

func RequireOwner(c *gin.Context, userID string) bool {
	claims, ok := ClaimsFrom(c)
	if !ok {
		abort(c, http.StatusUnauthorized, constants.ErrUnauthorized)
		return false
	}
	if CanBypassOwnership(c) || claims.Subject == userID {
		return true
	}
	abort(c, http.StatusForbidden, constants.ErrForbidden)
	return false
}

func abort(c *gin.Context, status int, message string) {
	c.AbortWithStatusJSON(status, gin.H{"error": message})
}
