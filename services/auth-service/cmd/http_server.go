package main

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
)

func (app *application) ServeHTTP(ctx context.Context) error {
	isProduction := app.config.HTTP.Mode == "production"

	if isProduction {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	router := app.routes()

	if !isProduction {
		pprof.Register(router)
	}

	server := &http.Server{
		Addr:         app.config.HTTP.Addr,
		Handler:      router,
		ReadTimeout:  app.config.HTTP.ReadTimeout,
		WriteTimeout: app.config.HTTP.WriteTimeout,
		IdleTimeout:  app.config.HTTP.IdleTimeout,
	}

	shutdown := make(chan error, 1)
	go func() {
		<-ctx.Done()
		stopCtx, cancel := context.WithTimeout(context.Background(), app.config.HTTP.ShutdownTimeout)
		defer cancel()
		shutdown <- server.Shutdown(stopCtx)
	}()

	app.logger.Info().Str("addr", server.Addr).Msg("starting HTTP server")
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return <-shutdown
}
