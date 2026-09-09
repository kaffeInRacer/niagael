package main

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (app *application) ServeHTTP(ctx context.Context) error {
	if app.config.HTTP.Mode == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	server := &http.Server{
		Addr:         app.config.HTTP.Addr,
		Handler:      app.routes(),
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
