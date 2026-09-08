package config

import (
	"fmt"
	"os"
	"strconv"

	"gopkg.in/yaml.v3"

	"kaffein/auth-service/utils/env"
)

func Load(path string) (*Config, error) {
	cfg := &Config{}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config %s: %w", path, err)
	}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parse config %s: %w", path, err)
	}

	cfg.HTTP.Addr = env.GetString("HTTP_ADDR", cfg.HTTP.Addr)
	cfg.Postgres.DSN = env.GetString("POSTGRES_DSN", cfg.Postgres.DSN)
	cfg.Redis.Addr = env.GetString("REDIS_ADDR", cfg.Redis.Addr)
	cfg.Redis.Password = env.GetString("REDIS_PASSWORD", cfg.Redis.Password)
	cfg.JWT.Issuer = env.GetString("JWT_ISSUER", cfg.JWT.Issuer)
	cfg.JWT.Secret = env.GetString("JWT_SECRET", cfg.JWT.Secret)
	if value, ok := os.LookupEnv("JWT_COOKIE_SECURE"); ok && value != "" {
		cookieSecure, err := strconv.ParseBool(value)
		if err != nil {
			return nil, fmt.Errorf("parse JWT_COOKIE_SECURE: %w", err)
		}
		cfg.JWT.CookieSecure = cookieSecure
	}
	return cfg, nil
}
