package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/openlankapay/openlankapay/pkg/observability"
	"github.com/openlankapay/openlankapay/pkg/ratelimit"
	"github.com/openlankapay/openlankapay/services/gateway/internal/handler"
	"github.com/openlankapay/openlankapay/services/gateway/internal/middleware"
	"github.com/openlankapay/openlankapay/services/gateway/internal/proxy"
)

func main() {
	logger := observability.NewLogger("gateway", getEnv("LOG_LEVEL", "info"))

	// Service proxy configuration
	serviceProxy := proxy.NewServiceProxy(proxy.Config{
		MerchantServiceURL:     getEnv("MERCHANT_SERVICE_URL", "http://localhost:8082"),
		PaymentServiceURL:      getEnv("PAYMENT_SERVICE_URL", "http://localhost:8081"),
		ExchangeServiceURL:     getEnv("EXCHANGE_SERVICE_URL", "http://localhost:8085"),
		SettlementServiceURL:   getEnv("SETTLEMENT_SERVICE_URL", "http://localhost:8083"),
		WebhookServiceURL:      getEnv("WEBHOOK_SERVICE_URL", "http://localhost:8084"),
		SubscriptionServiceURL: getEnv("SUBSCRIPTION_SERVICE_URL", "http://localhost:8086"),
		NotificationServiceURL: getEnv("NOTIFICATION_SERVICE_URL", "http://localhost:8087"),
		AdminServiceURL:        getEnv("ADMIN_SERVICE_URL", "http://localhost:8088"),
		DirectDebitServiceURL:  getEnv("DIRECTDEBIT_SERVICE_URL", "http://localhost:8089"),
	})

	port := getEnv("PORT", "8080")

	// Rate limiter: prefer Redis, fall back to in-memory
	var rateLimiter middleware.RateLimiter
	redisURL := getEnv("REDIS_URL", "localhost:6379")
	redisClient := redis.NewClient(&redis.Options{Addr: redisURL})
	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		logger.Warn().Err(err).Msg("Redis unavailable, using in-memory rate limiter")
		rateLimiter = middleware.NewInMemoryRateLimiter(100)
	} else {
		logger.Info().Msg("using Redis-backed rate limiter")
		rateLimiter = ratelimit.NewRedisRateLimiter(redisClient, ratelimit.DefaultConfig())
	}

	// API key validator for HMAC-authenticated SDK routes
	merchantURL := getEnv("MERCHANT_SERVICE_URL", "http://localhost:8082")
	keyValidator := middleware.NewRemoteKeyValidator(merchantURL)

	cfg := handler.GatewayConfig{
		JWTSecret:          getEnv("JWT_SECRET", "dev-jwt-secret-change-in-production-min32chars"),
		ServiceProxy:       serviceProxy,
		RateLimitPerMinute: 100,
		RateLimiter:        rateLimiter,
		KeyValidator:       keyValidator,
		GatewayPort:        port,
		PlatformFeePct:     getEnv("PLATFORM_FEE_PCT", "1.5"),
		ExchangeFeePct:     getEnv("EXCHANGE_FEE_PCT", "0.5"),
	}

	router := handler.NewGatewayRouter(cfg)
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
		logger.Info().Msg("shutting down gateway...")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
	}()

	logger.Info().Str("port", port).Msg("gateway started")
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
