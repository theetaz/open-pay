package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/openlankapay/openlankapay/pkg/observability"
	"github.com/openlankapay/openlankapay/services/gateway/internal/handler"
	"github.com/openlankapay/openlankapay/services/gateway/internal/proxy"
)

func main() {
	logger := observability.NewLogger("gateway", getEnv("LOG_LEVEL", "info"))

	// Service proxy configuration
	serviceProxy := proxy.NewServiceProxy(proxy.Config{
		MerchantServiceURL: getEnv("MERCHANT_SERVICE_URL", "http://localhost:8082"),
		PaymentServiceURL:  getEnv("PAYMENT_SERVICE_URL", "http://localhost:8081"),
		ExchangeServiceURL: getEnv("EXCHANGE_SERVICE_URL", "http://localhost:8085"),
	})

	cfg := handler.GatewayConfig{
		JWTSecret:          getEnv("JWT_SECRET", "dev-jwt-secret-change-in-production-min32chars"),
		ServiceProxy:       serviceProxy,
		RateLimitPerMinute: 100,
	}

	router := handler.NewGatewayRouter(cfg)

	port := getEnv("PORT", "8080")
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
