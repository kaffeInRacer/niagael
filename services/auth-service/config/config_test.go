package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadJWTConfigCookieSecureFromEnvironment(t *testing.T) {
	t.Setenv("JWT_COOKIE_SECURE", "true")
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte("jwt:\n  cookie_secure: false\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.JWT.CookieSecure {
		t.Fatal("JWT_COOKIE_SECURE did not override YAML configuration")
	}
}
