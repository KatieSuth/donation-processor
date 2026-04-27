// Command server is the boilerplate HTTP API: loads configuration from the environment,
// runs database migrations, and serves Gin routes.
package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/KatieSuth/donation-processor/backend/internal/handler"
	"github.com/KatieSuth/donation-processor/backend/internal/logger"
	"github.com/KatieSuth/donation-processor/backend/internal/middleware"
	"github.com/KatieSuth/donation-processor/backend/internal/store"
	"github.com/KatieSuth/donation-processor/backend/scripts"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

func fatalExit(msg string, args ...any) {
	slog.Error(msg, args...)
	os.Exit(1)
}

// configureLogger sets process-wide slog defaults from GIN_MODE.
// release mode uses production-safe verbosity; all other modes use dev/test verbosity.
func configureLogger(ginEnv string) {
	options := &slog.HandlerOptions{Level: slog.LevelDebug, AddSource: true}
	if ginEnv == gin.ReleaseMode {
		options = &slog.HandlerOptions{Level: slog.LevelInfo}
	}

	base := slog.NewJSONHandler(os.Stdout, options)
	slog.SetDefault(slog.New(logger.New(base)))
}

func main() {
	ginEnv := os.Getenv("GIN_MODE")
	configureLogger(ginEnv)

	switch ginEnv {
	case gin.ReleaseMode:
		gin.SetMode(gin.ReleaseMode)
	case gin.TestMode:
		gin.SetMode(gin.TestMode)
	default:
		gin.SetMode(gin.DebugMode)
	}

	/*** ENVIRONMENT VARIABLES ***/
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Database
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		fatalExit("DATABASE_URL is required")
	}

	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "https://donation-processor.localhost"
	}

	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		fatalExit("db connect failed", "error", err)
	}
	defer pool.Close()

	// Run migrations

	// wrap the existing pgxpool in a *sql.DB interface without opening a second connection
	sqlDB := stdlib.OpenDBFromPool(pool)
	defer sqlDB.Close()

	if err := goose.SetDialect("postgres"); err != nil {
		fatalExit("failed to set goose dialect", "error", err)
	}

	goose.SetBaseFS(nil) // use OS filesystem
	if err := goose.Up(sqlDB, "sql/migrations"); err != nil {
		fatalExit("failed to run migrations", "error", err)
	}
	if gin.Mode() == gin.DebugMode {
		if err := scripts.SeedTestData(context.Background(), pool); err != nil {
			fatalExit("failed to seed debug test data", "error", err)
		}
	}

	// Handlers
	s := store.NewPostgresStore(pool)
	h := handler.New(frontendURL, s)

	r := gin.New()
	r.Use(middleware.RequestID(), middleware.Recovery(), middleware.RequestLogger())
	r.SetTrustedProxies([]string{"172.30.0.0/16"})
	r.GET("/health", h.Health)

	donations := r.Group("/donations")
	{
		donations.GET("", h.ListDonations)
		donations.GET("/:uuid", h.GetDonation)
		donations.POST("", h.CreateDonation)
		donations.PATCH("/:uuid/status", h.UpdateDonationStatus)
	}

	slog.Info("API listening", "port", port)
	if err := r.Run(":" + port); err != nil {
		fatalExit("server error", "error", err)
	}
}
