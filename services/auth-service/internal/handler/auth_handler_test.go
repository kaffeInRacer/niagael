package handler

import (
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"kaffein/auth-service/config"
	"kaffein/auth-service/internal/dto"
)

func TestSetTokenCookiesAndHideTokensFromJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := &authHandler{jwt: config.JWTConfig{AccessTTL: 15 * time.Minute, RefreshTTL: 7 * 24 * time.Hour, CookieSecure: true}}
	response := &dto.TokenResponse{AccessToken: "access-secret", RefreshToken: "refresh-secret", TokenType: "Bearer", ExpiresIn: 900}
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)

	h.setTokenCookies(c, response)
	c.JSON(200, response)

	var body map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if _, ok := body["access_token"]; ok {
		t.Fatal("access token exposed in JSON")
	}
	if _, ok := body["refresh_token"]; ok {
		t.Fatal("refresh token exposed in JSON")
	}

	cookies := recorder.Result().Cookies()
	if len(cookies) != 2 {
		t.Fatalf("got %d cookies, want 2", len(cookies))
	}
	assertCookie := func(cookieName string, maxAge int) {
		t.Helper()
		for _, cookie := range cookies {
			if cookie.Name != cookieName {
				continue
			}
			if !cookie.HttpOnly || !cookie.Secure || cookie.SameSite != 2 || cookie.Path != "/" || cookie.MaxAge != maxAge {
				t.Fatalf("unexpected %s cookie: %+v", cookieName, cookie)
			}
			return
		}
		t.Fatalf("cookie %s not found", cookieName)
	}
	assertCookie(accessTokenCookie, 900)
	assertCookie(refreshTokenCookie, 604800)
}

func TestClearTokenCookies(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := &authHandler{}
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)

	h.clearTokenCookies(c)

	for _, cookie := range recorder.Result().Cookies() {
		if cookie.MaxAge != -1 || cookie.Value != "" || !cookie.HttpOnly || cookie.Path != "/" {
			t.Fatalf("cookie not cleared: %+v", cookie)
		}
	}
}
