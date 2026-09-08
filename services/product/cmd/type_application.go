package main

import (
	"kaffein/product-service/config"
	"kaffein/product-service/internal/auth"
	grpcclient "kaffein/product-service/pkg/grpc/client"
	"kaffein/product-service/pkg/rdb"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
)

type application struct {
	config        *config.Config
	logger        zerolog.Logger
	pgx           *pgxpool.Pool
	cache         *rdb.Cache
	pricingClient *grpcclient.DynamicPricingClient
	auth          *auth.Service
	wg            *sync.WaitGroup
}
