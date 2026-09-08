package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"kaffein/auth-service/config"
	"kaffein/auth-service/internal/interfaces/IRepository"
	"kaffein/auth-service/internal/token"
)

type authSessionRepository struct{}

func (authSessionRepository) Save(context.Context, uuid.UUID, IRepository.Session, time.Duration) error {
	return nil
}
func (authSessionRepository) Get(context.Context, uuid.UUID) (*IRepository.Session, error) {
	return nil, nil
}
func (authSessionRepository) Revoke(context.Context, uuid.UUID) error { return nil }
func (authSessionRepository) RevokeWithTTL(context.Context, uuid.UUID, time.Duration) error {
	return nil
}
func (authSessionRepository) IsRevoked(context.Context, uuid.UUID) (bool, error) { return false, nil }
func (authSessionRepository) RevokeUser(context.Context, uuid.UUID) error        { return nil }

func TestAuthAcceptsAccessTokenCookie(t *testing.T) {
	gin.SetMode(gin.TestMode)
	manager := token.NewManager(config.JWTConfig{Issuer: "test", Secret: "secret", AccessTTL: time.Minute, RefreshTTL: time.Hour})
	access, _, _, _, err := manager.Pair(uuid.New(), "tenant")
	if err != nil {
		t.Fatal(err)
	}

	router := gin.New()
	router.GET("/protected", Auth(manager, authSessionRepository{}), func(c *gin.Context) { c.Status(http.StatusNoContent) })
	request := httptest.NewRequest(http.MethodGet, "/protected", nil)
	request.AddCookie(&http.Cookie{Name: "access_token", Value: access})
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("got status %d, want %d", recorder.Code, http.StatusNoContent)
	}
}
