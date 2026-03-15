package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/openlankapay/openlankapay/services/exchange/internal/domain"
	"github.com/shopspring/decimal"
)

// ExchangeServiceInterface defines operations the handler depends on.
type ExchangeServiceInterface interface {
	GetActiveRate(ctx context.Context, base, quote string) (*domain.ExchangeRate, error)
	ConvertLKRToUSDT(ctx context.Context, lkrAmount decimal.Decimal) (decimal.Decimal, *domain.RateSnapshot, error)
}

// ExchangeHandler handles HTTP requests for exchange rate operations.
type ExchangeHandler struct {
	svc ExchangeServiceInterface
}

// NewExchangeHandler creates a new ExchangeHandler.
func NewExchangeHandler(svc ExchangeServiceInterface) *ExchangeHandler {
	return &ExchangeHandler{svc: svc}
}

// RegisterRoutes adds exchange rate routes to a chi router.
func (h *ExchangeHandler) RegisterRoutes(r chi.Router) {
	r.Get("/v1/exchange-rates/active", h.GetActiveRate)
}

// GetActiveRate handles GET /v1/exchange-rates/active?base=USDT&quote=LKR
func (h *ExchangeHandler) GetActiveRate(w http.ResponseWriter, r *http.Request) {
	base := r.URL.Query().Get("base")
	quote := r.URL.Query().Get("quote")

	if base == "" {
		base = "USDT"
	}
	if quote == "" {
		quote = "LKR"
	}

	rate, err := h.svc.GetActiveRate(r.Context(), base, quote)
	if err != nil {
		if errors.Is(err, domain.ErrRateNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]any{
				"error": map[string]string{"code": "NOT_FOUND", "message": "rate not found"},
			})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]any{
			"error": map[string]string{"code": "INTERNAL_ERROR", "message": "failed to get rate"},
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"data": map[string]any{
			"baseCurrency":  rate.BaseCurrency,
			"quoteCurrency": rate.QuoteCurrency,
			"rate":          rate.Rate.String(),
			"source":        rate.Source,
			"fetchedAt":     rate.FetchedAt.Format(time.RFC3339),
		},
	})
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}
