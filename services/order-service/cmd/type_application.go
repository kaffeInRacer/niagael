package main

import (
	"kaffein/order-service/config"
	"kaffein/order-service/internal/auth"
	grpcclient "kaffein/order-service/pkg/grpc/client"
	"kaffein/order-service/pkg/redis"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
	"sync"
)

type application struct {
	config        config.Config
	logger        zerolog.Logger
	pgx           *pgxpool.Pool
	redis         *redis.Redis
	productClient *grpcclient.ProductClient
	pricingClient *grpcclient.DynamicPricingClient
	auth          *auth.Service
	wg            *sync.WaitGroup
}
