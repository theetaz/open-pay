package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/openlankapay/openlankapay/pkg/audit"
	"github.com/openlankapay/openlankapay/pkg/database"
	"github.com/openlankapay/openlankapay/pkg/observability"
	pgadapter "github.com/openlankapay/openlankapay/services/merchant/internal/adapter/postgres"
	"github.com/openlankapay/openlankapay/services/merchant/internal/handler"
	"github.com/openlankapay/openlankapay/services/merchant/internal/service"
)

func main() {
	logger := observability.NewLogger("merchant", getEnv("LOG_LEVEL", "info"))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Database
	dbURL := getEnv("DATABASE_URL", "postgres://olp:olp_dev_password@localhost:5433/merchant_db?sslmode=disable")
	pool, err := database.NewPool(ctx, database.DefaultConfig(dbURL), logger)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to connect to database")
	}
	defer pool.Close()

	// JWT secret
	jwtSecret := getEnv("JWT_SECRET", "dev-jwt-secret-change-in-production-min32chars")

	// Repositories
	merchantRepo := pgadapter.NewMerchantRepository(pool)
	apiKeyRepo := pgadapter.NewAPIKeyRepository(pool)
	userRepo := pgadapter.NewUserRepository(pool)
	branchRepo := pgadapter.NewBranchRepository(pool)
	paymentLinkRepo := pgadapter.NewPaymentLinkRepository(pool)

	// Event publisher (noop for now)
	eventPub := &noopPublisher{}

	// Service
	svc := service.NewMerchantService(merchantRepo, apiKeyRepo, userRepo, eventPub, jwtSecret)

	// File upload handler (MinIO)
	minioEndpoint := getEnv("MINIO_ENDPOINT", "localhost:9000")
	minioAccessKey := getEnv("MINIO_ACCESS_KEY", "minioadmin")
	minioSecretKey := getEnv("MINIO_SECRET_KEY", "minioadmin123")
	minioBucket := getEnv("MINIO_BUCKET", "kyc-documents")
	var uploadHandler *handler.FileUploadHandler
	uploadHandler, err = handler.NewFileUploadHandler(handler.FileUploadConfig{
		Endpoint:  minioEndpoint,
		AccessKey: minioAccessKey,
		SecretKey: minioSecretKey,
		Bucket:    minioBucket,
		UseSSL:    false,
	})
	if err != nil {
		logger.Warn().Err(err).Msg("MinIO not available, file uploads disabled")
		uploadHandler = nil
	} else {
		logger.Info().Msg("file upload handler initialized with MinIO")
	}

	// Audit log client
	adminServiceURL := getEnv("ADMIN_SERVICE_URL", "http://localhost:8088")
	auditClient := audit.NewClient(adminServiceURL)

	// HTTP Handler
	h := handler.NewMerchantHandler(svc, jwtSecret, auditClient)
	router := handler.NewRouter(h, branchRepo, paymentLinkRepo, uploadHandler)

	// Server
	port := getEnv("PORT", "8082")
	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		logger.Info().Msg("shutting down...")
		cancel()

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()
		_ = srv.Shutdown(shutdownCtx)
	}()

	logger.Info().Str("port", port).Msg("merchant service started")
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Fatal().Err(err).Msg("server error")
	}
}

type noopPublisher struct{}

func (p *noopPublisher) Publish(_ context.Context, _ string, _ any) error {
	return nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
