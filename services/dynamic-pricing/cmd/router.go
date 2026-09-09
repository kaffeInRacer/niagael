package main

import (
	"context"
	"kaffein/dynamic-pricing-service/internal/handler"
	"kaffein/dynamic-pricing-service/internal/repository"
	"kaffein/dynamic-pricing-service/internal/usecase"
	"kaffein/dynamic-pricing-service/pkg/postgresql"
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
			ExposeHeaders:    []string{"Content-Length"},
			AllowCredentials: true,
			MaxAge:           12 * time.Hour,
		}),
		app.requestLogger(),
	)

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": true, "message": "ok"})
	})

	store := postgresql.NewStore(app.pgx)

	flashSaleRepo := repository.NewFlashSaleRepository(store)
	flashSaleUsageRepo := repository.NewFlashSaleUsageRepository(store)
	flashSaleUseCase := usecase.NewFlashSaleUseCase(flashSaleRepo, flashSaleUsageRepo)
	handler.NewFlashSaleHandler(flashSaleUseCase, app.logger, r, app.auth)

	promoRepo := repository.NewPromoRepository(store)
	promoUsageRepo := repository.NewPromoUsageRepository(store)
	promoUseCase := usecase.NewPromoUseCase(promoRepo, promoUsageRepo)
	handler.NewPromoHandler(promoUseCase, app.logger, r, app.auth)

	allocationRepo := repository.NewPricingAllocationRepository(store)
	app.startGrpcServer(ctx, flashSaleRepo, flashSaleUsageRepo, promoRepo, promoUsageRepo, allocationRepo)
	return r
}
