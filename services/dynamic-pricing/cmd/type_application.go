package main

import (
	"kaffein/dynamic-pricing-service/config"
	"kaffein/dynamic-pricing-service/internal/auth"
	"kaffein/dynamic-pricing-service/pkg/redis"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
)

type application struct {
	config *config.Config
	logger zerolog.Logger
	pgx    *pgxpool.Pool
	cache  *redis.Cache
	auth   *auth.Service
	wg     *sync.WaitGroup
}
