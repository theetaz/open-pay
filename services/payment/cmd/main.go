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
	"github.com/openlankapay/openlankapay/pkg/fraud"
	"github.com/openlankapay/openlankapay/pkg/observability"
	exchangeclient "github.com/openlankapay/openlankapay/services/payment/internal/adapter/exchange"
	merchantclient "github.com/openlankapay/openlankapay/services/payment/internal/adapter/merchant"
	pgadapter "github.com/openlankapay/openlankapay/services/payment/internal/adapter/postgres"
	"github.com/openlankapay/openlankapay/services/payment/internal/adapter/provider"
	"github.com/openlankapay/openlankapay/services/payment/internal/domain"
	"github.com/openlankapay/openlankapay/services/payment/internal/handler"
	"github.com/openlankapay/openlankapay/services/payment/internal/service"
)

func main() {
	logger := observability.NewLogger("payment", getEnv("LOG_LEVEL", "info"))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Database
	dbURL := getEnv("DATABASE_URL", "postgres://olp:olp_dev_password@localhost:5433/payment_db?sslmode=disable")
	pool, err := database.NewPool(ctx, database.DefaultConfig(dbURL), logger)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to connect to database")
	}
	defer pool.Close()

	// JWT secret
	jwtSecret := getEnv("JWT_SECRET", "dev-jwt-secret-change-in-production-min32chars")

	// Providers
	mockProv := provider.NewMockProvider()
	providers := map[string]domain.PaymentProvider{
		"TEST": mockProv,
	}
	logger.Info().Msg("registered provider: TEST")

	if apiKey := os.Getenv("BYBIT_API_KEY"); apiKey != "" {
		bp := provider.NewBybitProvider(provider.BybitConfig{
			BaseURL:   getEnv("BYBIT_BASE_URL", "https://api-testnet.bybit.com"),
			APIKey:    apiKey,
			APISecret: os.Getenv("BYBIT_API_SECRET"),
		})
		providers["BYBIT"] = bp
		logger.Info().Msg("registered provider: BYBIT")
	}

	if apiKey := os.Getenv("BINANCE_API_KEY"); apiKey != "" {
		bp := provider.NewBinanceProvider(provider.BinanceConfig{
			BaseURL:   getEnv("BINANCE_BASE_URL", "https://bpay.binanceapi.com"),
			APIKey:    apiKey,
			APISecret: os.Getenv("BINANCE_API_SECRET"),
		})
		providers["BINANCE"] = bp
		logger.Info().Msg("registered provider: BINANCE")
	}

	if apiKey := os.Getenv("KUCOIN_API_KEY"); apiKey != "" {
		kp := provider.NewKuCoinProvider(provider.KuCoinConfig{
			BaseURL:    getEnv("KUCOIN_BASE_URL", "https://api.kucoin.com"),
			APIKey:     apiKey,
			APISecret:  os.Getenv("KUCOIN_API_SECRET"),
			Passphrase: os.Getenv("KUCOIN_PASSPHRASE"),
		})
		providers["KUCOIN"] = kp
		logger.Info().Msg("registered provider: KUCOIN")
	}

	// Repository
	paymentRepo := pgadapter.NewPaymentRepository(pool)

	// Exchange client
	exchangeURL := getEnv("EXCHANGE_SERVICE_URL", "http://localhost:8085")
	exchClient := exchangeclient.NewClient(exchangeURL)

	// Event publisher (noop for now)
	eventPub := &noopPublisher{}

	// Merchant client (for incrementing payment link usage)
	merchantServiceURL := getEnv("MERCHANT_SERVICE_URL", "http://localhost:8082")
	merchClient := merchantclient.NewClient(merchantServiceURL)

	// Audit client
	adminServiceURL := getEnv("ADMIN_SERVICE_URL", "http://localhost:8088")
	auditClient := audit.NewClient(adminServiceURL)

	// Fraud engine
	fraudEngine := fraud.NewEngine()
	logger.Info().Msg("fraud detection engine initialized")

	// Service
	svc := service.NewPaymentService(paymentRepo, providers, exchClient, eventPub, merchClient)
	svc.SetFraudEngine(fraudEngine)

	// HTTP Handler
	h := handler.NewPaymentHandler(svc, mockProv, auditClient)
	router := handler.NewRouter(h, jwtSecret)

	// Server
	port := getEnv("PORT", "8081")
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
		logger.Info().Msg("shutting down payment service...")
		cancel()

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()
		_ = srv.Shutdown(shutdownCtx)
	}()

	logger.Info().Str("port", port).Msg("payment service started")
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
