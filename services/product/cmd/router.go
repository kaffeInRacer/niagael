package main

import (
	"context"
	"kaffein/product-service/internal/handler"
	"kaffein/product-service/internal/repository"
	"kaffein/product-service/internal/usecase"
	minioStorage "kaffein/product-service/pkg/minio/storage"
	"kaffein/product-service/pkg/postgresql"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func (app *application) routes(ctx context.Context) *gin.Engine {
	r := gin.New()
	r.Use(
		gin.Recovery(),
		cors.New(cors.Config{
			AllowOrigins:     []string{"http://localhost:3000", "http://127.0.0.1:3000"},
			AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
			AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With"},
			AllowCredentials: true,
			MaxAge:           12 * time.Hour,
		}),
		app.requestLogger(),
	)

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": true, "message": "ok"})
	})

	store := postgresql.NewStore(app.pgx)

	storage, err := minioStorage.New(app.config)
	if err != nil {
		app.logger.Fatal().Err(err).Msg("failed to initialize minio storage")
	}

	categoryRepo := repository.NewCategoryRepository(store)
	productRepo := repository.NewProductRepository(store)
	productImageRepo := repository.NewProductImageRepository(store)
	variantRepo := repository.NewVariantRepository(store)
	stockRepo := repository.NewStockRepository(store)

	categoryUseCase := usecase.NewCategoryUseCase(categoryRepo)
	productUseCase := usecase.NewProductUseCase(productRepo, productImageRepo, variantRepo, stockRepo)
	productImageUseCase := usecase.NewProductImageUseCase(productImageRepo)
	variantUseCase := usecase.NewVariantUseCase(variantRepo)

	productImageHandler := handler.NewProductImageHandler(productImageUseCase, storage, app.logger)
	handler.NewCategoryHandler(categoryUseCase, app.logger, r, app.auth)
	handler.NewVariantHandler(variantUseCase, app.logger, r, app.auth)
	handler.NewProductHandler(productUseCase, app.pricingClient, app.logger, r, productImageHandler, app.auth, repository.NewFlashSaleProjectionRepository(app.pgx))

	app.startGrpcServer(ctx, productUseCase)
	return r
}
