package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"server/internal/core/config"
	"server/internal/core/logger"
	"server/internal/core/repository/mysql/pool/stdsql"
	"server/internal/core/transport/http/middleware"
	"server/internal/core/transport/http/server"
	authgoogle "server/internal/features/auth/repository/google"
	authmysql "server/internal/features/auth/repository/mysql"
	authservice "server/internal/features/auth/service"
	authhttp "server/internal/features/auth/transport/http"
	usersmysql "server/internal/features/users/repository/mysql"
	usersservice "server/internal/features/users/service"
	usershttp "server/internal/features/users/transport/http"
)

func main() {
	if err := run(); err != nil {
		slog.Error("application stopped with error", "err", err)
		os.Exit(1)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	log, logCloser, err := logger.New(logger.Config{
		Level:      cfg.Logger.Level,
		FilePath:   cfg.Logger.FilePath,
		MaxSizeMB:  cfg.Logger.MaxSizeMB,
		MaxBackups: cfg.Logger.MaxBackups,
		MaxAgeDays: cfg.Logger.MaxAgeDays,
		Compress:   cfg.Logger.Compress,
	})
	if err != nil {
		return fmt.Errorf("init logger: %w", err)
	}
	defer logCloser.Close()
	slog.SetDefault(log)

	dbPool, err := stdsql.New(stdsql.Config{
		DSN:             cfg.DB.DSN,
		OpTimeout:       cfg.DB.OpTimeout,
		ConnMaxLifetime: cfg.DB.ConnMaxLifetime,
		MaxOpenConns:    cfg.DB.MaxOpenConns,
		MaxIdleConns:    cfg.DB.MaxIdleConns,
	})
	if err != nil {
		return fmt.Errorf("init mysql pool: %w", err)
	}
	defer dbPool.Close()

	pingCtx, cancelPing := context.WithTimeout(ctx, dbPool.OpTimeout())
	defer cancelPing()
	if err := dbPool.Ping(pingCtx); err != nil {
		return fmt.Errorf("ping mysql: %w", err)
	}

	srv := server.New(
		server.Config{
			Port:            cfg.Server.Port,
			ShutdownTimeout: cfg.Server.ShutdownTimeout,
			Title:           "Nodero API",
			Version:         "1.0.0",
		},
		log,
		middleware.RequestID(),
		middleware.Logger(log),
		middleware.Trace(),
		middleware.CORS(cfg.Server.AllowedOrigins),
		middleware.Recover(),
	)

	v1 := srv.APIVersionGroup("/api/v1")

	usersRepo := usersmysql.NewRepository(dbPool)
	usersSvc := usersservice.NewService(usersRepo)
	usersHandler := usershttp.NewHandler(usersSvc)
	usersHandler.Register(v1)

	authRepo := authmysql.NewRepository(dbPool)
	googleProvider := authgoogle.NewProvider(authgoogle.Config{
		ClientID:     cfg.OAuth.Google.ClientID,
		ClientSecret: cfg.OAuth.Google.ClientSecret,
		RedirectURL:  fmt.Sprintf("%s/auth/google/callback", cfg.OAuth.FrontendURL),
	})
	authSvc := authservice.NewService(dbPool, authRepo, usersSvc, googleProvider)
	authMiddleware := middleware.NewAuth(v1, authSvc)
	authHandler := authhttp.NewHandler(authSvc)
	authHandler.Register(v1, authMiddleware)

	return srv.Run(ctx)
}
