package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"

	"github.com/KozhabergenovNurzhan/E-commerce/config"
	"github.com/KozhabergenovNurzhan/E-commerce/internal/auth"
	"github.com/KozhabergenovNurzhan/E-commerce/internal/handler"
	"github.com/KozhabergenovNurzhan/E-commerce/internal/repository"
	"github.com/KozhabergenovNurzhan/E-commerce/internal/server"
	"github.com/KozhabergenovNurzhan/E-commerce/internal/service"
	"github.com/KozhabergenovNurzhan/E-commerce/pkg/logger"
)

func main() {
	cfg := config.Load()
	log := logger.New("ecommerce", cfg.LogLevel)
	slog.SetDefault(log)

	db := mustConnectDB(cfg.DB, log)
	defer db.Close()

	runMigrations(cfg.DB, log)

	authMgr := auth.NewJWTManager(cfg.JWT.Secret, cfg.JWT.AccessTTL)

	userRepo    := repository.NewUserRepository(db)
	productRepo := repository.NewProductRepository(db)
	orderRepo   := repository.NewOrderRepository(db)
	tokenRepo   := repository.NewTokenRepository(db)
	cartRepo    := repository.NewCartRepository(db)

	svc := service.NewServices(
		service.NewUserService(userRepo),
		service.NewProductService(productRepo),
		service.NewOrderService(orderRepo, productRepo),
		service.NewTokenService(tokenRepo, userRepo, authMgr, cfg.JWT.RefreshTTL),
		service.NewCartService(cartRepo, productRepo),
	)

	h := handler.NewHandler(svc, authMgr, log)
	srv := server.New(cfg.Port, h.InitRoutes())

	go func() {
		log.Info("server starting", slog.String("port", cfg.Port))
		if err := srv.Run(); err != nil && err != http.ErrServerClosed {
			log.Error("server error", slog.String("err", err.Error()))
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Error("shutdown error", slog.String("err", err.Error()))
	}
	log.Info("server stopped")
}

func mustConnectDB(cfg config.DBConfig, log *slog.Logger) *sqlx.DB {
	db, err := sqlx.Connect("pgx", cfg.DSN())
	if err != nil {
		log.Error("failed to connect to postgres", slog.String("err", err.Error()))
		os.Exit(1)
	}
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)
	log.Info("connected to postgres")
	return db
}

func runMigrations(cfg config.DBConfig, log *slog.Logger) {
	m, err := migrate.New("file://migrations", cfg.MigrateURL())
	if err != nil {
		log.Error("failed to init migrations", slog.String("err", err.Error()))
		os.Exit(1)
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Error("migration failed", slog.String("err", err.Error()))
		os.Exit(1)
	}
	log.Info("migrations applied")
}
