package handler

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"kaffein/auth-service/config"
	"kaffein/auth-service/internal/dto"
	"kaffein/auth-service/internal/interfaces/IUseCase"
	"kaffein/auth-service/internal/middleware"
	"kaffein/auth-service/utils/constants"
)

type authHandler struct {
	usecase IUseCase.AuthUseCase
	logger  zerolog.Logger
	jwt     config.JWTConfig
}

const (
	accessTokenCookie  = "access_token"
	refreshTokenCookie = "refresh_token"
)

func NewAuthHandler(usecase IUseCase.AuthUseCase, logger zerolog.Logger, engine *gin.Engine, auth gin.HandlerFunc, jwt config.JWTConfig) *authHandler {
	h := &authHandler{usecase: usecase, logger: logger, jwt: jwt}
	routes := engine.Group("/auth")
	routes.POST("/register", h.register)
	routes.POST("/login", h.login)
	routes.POST("/refresh", h.refresh)
	routes.POST("/logout", auth, h.logout)
	routes.GET("/me", auth, h.me)
	return h
}

func (h *authHandler) register(c *gin.Context) {
	var req dto.RegisterRequest
	if !bind(c, &req) {
		return
	}
	user, err := h.usecase.Register(c.Request.Context(), req)
	if err != nil {
		h.respondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, user)
}

func (h *authHandler) login(c *gin.Context) {
	var req dto.LoginRequest
	if !bind(c, &req) {
		return
	}
	result, err := h.usecase.Login(c.Request.Context(), req)
	if err != nil {
		h.respondError(c, err)
		return
	}
	h.setTokenCookies(c, result)
	c.JSON(http.StatusOK, result)
}

func (h *authHandler) refresh(c *gin.Context) {
	raw, err := c.Cookie(refreshTokenCookie)
	if err != nil {
		parts := strings.Fields(c.GetHeader("Authorization"))
		if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
			raw = parts[1]
		} else {
			var req dto.RefreshRequest
			if !bind(c, &req) {
				return
			}
			raw = req.RefreshToken
		}
	}
	result, err := h.usecase.Refresh(c.Request.Context(), raw)
	if err != nil {
		h.respondError(c, err)
		return
	}
	h.setTokenCookies(c, result)
	c.JSON(http.StatusOK, result)
}

func (h *authHandler) logout(c *gin.Context) {
	h.clearTokenCookies(c)
	err := h.usecase.Logout(c.Request.Context(), c.MustGet(middleware.SessionIDKey).(uuid.UUID), c.MustGet(middleware.ExpirationKey).(int64))
	if err != nil {
		h.logger.Error().Err(err).Msg("session revoke failed during logout")
	}
	c.Status(http.StatusNoContent)
}

func (h *authHandler) setTokenCookies(c *gin.Context, result *dto.TokenResponse) {
	h.setCookie(c, accessTokenCookie, result.AccessToken, h.jwt.AccessTTL)
	h.setCookie(c, refreshTokenCookie, result.RefreshToken, h.jwt.RefreshTTL)
}

func (h *authHandler) clearTokenCookies(c *gin.Context) {
	h.setCookie(c, accessTokenCookie, "", -time.Second)
	h.setCookie(c, refreshTokenCookie, "", -time.Second)
}

func (h *authHandler) setCookie(c *gin.Context, name, value string, ttl time.Duration) {
	maxAge := int(ttl.Seconds())
	if ttl < 0 {
		maxAge = -1
	}
	http.SetCookie(c.Writer, &http.Cookie{
		Name: name, Value: value, Path: "/", MaxAge: maxAge,
		HttpOnly: true, Secure: h.jwt.CookieSecure, SameSite: http.SameSiteLaxMode,
	})
}

func (h *authHandler) me(c *gin.Context) {
	user, err := h.usecase.Me(c.Request.Context(), c.MustGet(middleware.UserIDKey).(uuid.UUID))
	if err != nil {
		h.respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, user)
}

func bind(c *gin.Context, value any) bool {
	if err := c.ShouldBindJSON(value); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return false
	}
	return true
}

func (h *authHandler) respondError(c *gin.Context, err error) {
	switch err.Error() {
	case constants.ErrEmailExists:
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case constants.ErrInvalidCredentials, constants.ErrInvalidToken, constants.ErrSessionNotFound:
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
	case constants.ErrUserDisabled:
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
	case constants.ErrUserNotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	default:
		h.logger.Error().Err(err).Msg("auth request failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
}
