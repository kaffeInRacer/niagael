package auth

import (
	"net/http"
	"testing"
)

func TestAccessToken(t *testing.T) {
	tests := []struct {
		name       string
		cookie     *http.Cookie
		header     string
		wantToken  string
		wantExists bool
	}{
		{name: "cookie preferred", cookie: &http.Cookie{Name: "access_token", Value: "cookie-token"}, header: "Bearer header-token", wantToken: "cookie-token", wantExists: true},
		{name: "bearer fallback", header: "Bearer header-token", wantToken: "header-token", wantExists: true},
		{name: "bearer case insensitive", header: "bearer header-token", wantToken: "header-token", wantExists: true},
		{name: "empty cookie does not fall back", cookie: &http.Cookie{Name: "access_token", Value: ""}, header: "Bearer header-token"},
		{name: "missing credentials"},
		{name: "wrong authorization scheme", header: "Basic credentials"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, "/", nil)
			if err != nil {
				t.Fatal(err)
			}
			if tt.cookie != nil {
				req.AddCookie(tt.cookie)
			}
			if tt.header != "" {
				req.Header.Set("Authorization", tt.header)
			}

			token, exists := accessToken(req)
			if token != tt.wantToken || exists != tt.wantExists {
				t.Fatalf("accessToken() = (%q, %v), want (%q, %v)", token, exists, tt.wantToken, tt.wantExists)
			}
		})
	}
}
