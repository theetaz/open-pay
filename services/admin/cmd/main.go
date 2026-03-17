package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/openlankapay/openlankapay/pkg/database"
	"github.com/openlankapay/openlankapay/pkg/observability"
	pgadapter "github.com/openlankapay/openlankapay/services/admin/internal/adapter/postgres"
	"github.com/openlankapay/openlankapay/services/admin/internal/handler"
	"github.com/openlankapay/openlankapay/services/admin/internal/service"
)

func main() {
	logger := observability.NewLogger("admin", getEnv("LOG_LEVEL", "info"))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dbURL := getEnv("DATABASE_URL", "postgres://olp:olp_dev_password@localhost:5433/admin_db?sslmode=disable")
	pool, err := database.NewPool(ctx, database.DefaultConfig(dbURL), logger)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to connect to database")
	}
	defer pool.Close()

	jwtSecret := getEnv("JWT_SECRET", "dev-jwt-secret-change-in-production-min32chars")

	auditRepo := pgadapter.NewAuditRepository(pool)
	adminUserRepo := pgadapter.NewAdminUserRepository(pool)
	legalDocRepo := pgadapter.NewLegalDocumentRepository(pool)

	auditSvc := service.NewAuditService(auditRepo)
	authSvc := service.NewAdminAuthService(adminUserRepo, jwtSecret)

	h := handler.NewAdminHandler(auditSvc, authSvc, legalDocRepo)
	router := handler.NewRouter(h, jwtSecret)

	port := getEnv("PORT", "8088")
	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		logger.Info().Msg("shutting down admin service...")
		cancel()
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()
		_ = srv.Shutdown(shutdownCtx)
	}()

	logger.Info().Str("port", port).Msg("admin service started")
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Fatal().Err(err).Msg("server error")
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
