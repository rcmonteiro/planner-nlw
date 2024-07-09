package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"planner/internal/api/spec"
	"syscall"
	"time"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/phenpessoa/gutils/netutils/httputils"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	ctx := context.Background()
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt, os.Kill, syscall.SIGTERM, syscall.SIGKILL)
	defer cancel()

	if err := run(ctx); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
	fmt.Println("Goodbye!")
}

func run(ctx context.Context) error {
	cfg := zap.NewProductionConfig()
	cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder

	logger, err := cfg.Build()
	if err != nil {
		return err
	}
	logger = logger.Named("planner_app")
	defer logger.Sync()

	pool, err := pgxpool.New(ctx, fmt.Sprintf(
		"user=%s password=%s host=%s port=%s dbname=%s",
		os.Getenv("PLANNER_DB_USER"),
		os.Getenv("PLANNER_DB_PASSWORD"),
		os.Getenv("PLANNER_DB_HOST"),
		os.Getenv("PLANNER_DB_PORT"),
		os.Getenv("PLANNER_DB_NAME"),
	),
	)
	if err != nil {
		return err
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		return err
	}

	si := api.NewApi(pool, logger)
	r := chi.NewMux()
	r.Use(middleware.RequestID, middleware.Recoverer, httputils.ChiLogger(logger))
	r.Mount("/", spec.Handler(si))

	srv := &http.Server{
		Addr:         ":8080",
		Handler:      r,
		IdleTimeout:  time.Minute,
		ReadTimeout:  time.Second * 5,
		WriteTimeout: time.Second * 5,
	}

	defer func() {
		const timeout = time.Second * 30
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			logger.Error("failed to shutdown server", zap.Error(err))
		}
	}()

	errChan := make(chan error, 1)

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			errChan <- err
		}
	}()

	select {
	case <-ctx.Done():
		return nil
	case err := <-errChan:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	}
}
