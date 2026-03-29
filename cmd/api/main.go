package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/KozhabergenovNurzhan/E-commerce/internal/pkg/logger"
	"github.com/gin-gonic/gin"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"

	"github.com/KozhabergenovNurzhan/E-commerce/internal/auth"
	"github.com/KozhabergenovNurzhan/E-commerce/internal/cache"
	"github.com/KozhabergenovNurzhan/E-commerce/internal/config"
	"github.com/KozhabergenovNurzhan/E-commerce/internal/handler"
	"github.com/KozhabergenovNurzhan/E-commerce/internal/repository"
	"github.com/KozhabergenovNurzhan/E-commerce/internal/server"
	"github.com/KozhabergenovNurzhan/E-commerce/internal/service"
)

func main() {
	cfg := config.Load()

	log := logger.New("ecommerce", cfg.LogLevel, cfg.LogFormat)
	slog.SetDefault(log)

	router, err := buildApp(cfg, log)
	if err != nil {
		log.Error("failed to build app", slog.String("err", err.Error()))
		os.Exit(1)
	}

	srv := server.New(cfg.Port, router)
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

func buildApp(cfg *config.Config, log *slog.Logger) (*gin.Engine, error) {
	db, err := connectDB(cfg.DB, log)
	if err != nil {
		return nil, err
	}

	if err := runMigrations(cfg.DB, log); err != nil {
		return nil, err
	}

	authMgr := auth.NewJWTManager(cfg.JWT.Secret, cfg.JWT.AccessTTL)

	userRepo := repository.NewUserRepository(db)
	productRepo := repository.NewProductRepository(db)
	orderRepo := repository.NewOrderRepository(db)
	tokenRepo := repository.NewTokenRepository(db)
	cartRepo := repository.NewCartRepository(db)
	reviewRepo := repository.NewReviewRepository(db)
	addressRepo := repository.NewAddressRepository(db)

	redisClient := initRedisClient(cfg)

	var productCache *cache.RedisCache
	var idempotencyStore *cache.IdempotencyStore
	if redisClient != nil {
		productCache = cache.NewProductCache(redisClient)
		idempotencyStore = cache.NewIdempotencyStore(redisClient, 24*time.Hour)
	}

	svc := service.NewServices(
		service.NewUserService(userRepo),
		service.NewProductService(productRepo, productCache),
		service.NewOrderService(db, orderRepo, productRepo),
		service.NewTokenService(db, tokenRepo, userRepo, authMgr, cfg.JWT.RefreshTTL),
		service.NewCartService(cartRepo, productRepo),
		service.NewReviewService(reviewRepo, productRepo),
		service.NewAddressService(addressRepo, db),
	)

	h := handler.NewHandler(svc, authMgr, log, db, idempotencyStore)
	return h.InitRoutes(), nil
}

func connectDB(cfg config.DBConfig, log *slog.Logger) (*sqlx.DB, error) {
	db, err := sqlx.Connect("pgx", cfg.DSN())
	if err != nil {
		log.Error("failed to connect to postgres", slog.String("err", err.Error()))
		return nil, err
	}
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)
	log.Info("connected to postgres")
	return db, nil
}

func runMigrations(cfg config.DBConfig, log *slog.Logger) error {
	m, err := migrate.New("file://migrations", cfg.MigrateURL())
	if err != nil {
		log.Error("failed to init migrations", slog.String("err", err.Error()))
		return err
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Error("migration failed", slog.String("err", err.Error()))
		return err
	}
	log.Info("migrations applied")
	return nil
}

func initRedisClient(cfg *config.Config) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		slog.Warn("redis is unavailable, continuing without cache", "error", err.Error())
		_ = client.Close()
		return nil
	}

	slog.Info("Redis connected successfully")
	return client
}
