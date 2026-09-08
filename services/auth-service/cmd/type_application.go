package main

import (
	"github.com/jackc/pgx/v5/pgxpool"
	redislib "github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"kaffein/auth-service/config"
)

type application struct {
	config *config.Config
	logger zerolog.Logger
	pgx    *pgxpool.Pool
	redis  *redislib.Client
}
