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

	router := app.routes(ctx)

	if !isProduction {
		pprof.Register(router)
	}

	srv := &http.Server{
		Addr:         app.config.HTTP.Addr,
		Handler:      router,
		IdleTimeout:  app.config.HTTP.IdleTimeout,
		ReadTimeout:  app.config.HTTP.ReadTimeout,
		WriteTimeout: app.config.HTTP.WriteTimeout,
	}

	shutdownErr := make(chan error, 1)
	go func() {
		<-ctx.Done()
		app.logger.Info().Msg("shutting down http server")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), app.config.HTTP.ShutdownTimeout)
		defer cancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			shutdownErr <- err
		}
		close(shutdownErr)
	}()

	app.logger.Info().
		Str("addr", srv.Addr).
		Dur("read_timeout", app.config.HTTP.ReadTimeout).
		Dur("write_timeout", app.config.HTTP.WriteTimeout).
		Dur("idle_timeout", app.config.HTTP.IdleTimeout).
		Msg("starting HTTP server")

	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		app.logger.Err(err).Msg("ListenAndServe failed")
		return err
	}

	err := <-shutdownErr
	if err != nil {
		app.logger.Error().Err(err).Msg("shutdown failed")
		return err
	}

	app.logger.Info().Msg("HTTP server stopped")
	app.wg.Wait()
	return nil
}
