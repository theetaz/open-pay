package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/openlankapay/openlankapay/pkg/auth"
	"github.com/openlankapay/openlankapay/services/payment/internal/adapter/provider"
	"github.com/openlankapay/openlankapay/services/payment/internal/domain"
	"github.com/openlankapay/openlankapay/services/payment/internal/handler"
	"github.com/openlankapay/openlankapay/services/payment/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testJWTSecret = "test-jwt-secret-key-at-least-32-chars"

// --- Stub service ---

type stubPaymentService struct {
	payments map[uuid.UUID]*domain.Payment
	mockProv *provider.MockProvider
}

func newStubService() *stubPaymentService {
	return &stubPaymentService{
		payments: make(map[uuid.UUID]*domain.Payment),
		mockProv: provider.NewMockProvider(),
	}
}

func (s *stubPaymentService) CreatePayment(ctx context.Context, input service.CreatePaymentInput) (*domain.Payment, error) {
	prov := s.mockProv

	payment, err := domain.NewPayment(domain.CreatePaymentInput{
		MerchantID:      input.MerchantID,
		BranchID:        input.BranchID,
		Amount:          input.Amount,
		Currency:        input.Currency,
		Provider:        input.Provider,
		MerchantTradeNo: input.MerchantTradeNo,
		WebhookURL:      input.WebhookURL,
		CustomerEmail:   input.CustomerEmail,
	})
	if err != nil {
		return nil, err
	}

	payment.SetFees(decimal.NewFromFloat(0.5), decimal.NewFromFloat(1.5))

	provResp, err := prov.CreatePayment(ctx, domain.ProviderPaymentRequest{
		Amount:   payment.AmountUSDT.String(),
		Currency: "USDT",
		OrderID:  payment.ID.String(),
	})
	if err != nil {
		return nil, err
	}

	payment.ProviderPayID = provResp.ProviderPayID
	payment.QRContent = provResp.QRContent
	payment.CheckoutLink = provResp.CheckoutLink

	s.payments[payment.ID] = payment
	return payment, nil
}

func (s *stubPaymentService) GetPayment(_ context.Context, id uuid.UUID) (*domain.Payment, error) {
	p, ok := s.payments[id]
	if !ok {
		return nil, domain.ErrPaymentNotFound
	}
	return p, nil
}

func (s *stubPaymentService) ListPayments(_ context.Context, merchantID uuid.UUID, _ service.ListParams) ([]*domain.Payment, int, error) {
	var result []*domain.Payment
	for _, p := range s.payments {
		if p.MerchantID == merchantID {
			result = append(result, p)
		}
	}
	return result, len(result), nil
}

func (s *stubPaymentService) ExpireStalePayments(_ context.Context) (int, error) {
	return 0, nil
}

func (s *stubPaymentService) HandleProviderCallback(_ context.Context, paymentID uuid.UUID) error {
	p, ok := s.payments[paymentID]
	if !ok {
		return domain.ErrPaymentNotFound
	}

	status, err := s.mockProv.GetPaymentStatus(context.Background(), p.ProviderPayID)
	if err != nil {
		return err
	}

	if status.Status == domain.StatusPaid {
		_ = p.MarkPaid(status.TxHash)
	}
	return nil
}

func authToken(merchantID uuid.UUID) string {
	token, _ := auth.GenerateToken(uuid.New(), merchantID, "ADMIN", nil, testJWTSecret, time.Hour)
	return token
}

func TestCreatePaymentHandler(t *testing.T) {
	svc := newStubService()
	h := handler.NewPaymentHandler(svc, svc.mockProv)
	router := handler.NewRouter(h, testJWTSecret)
	merchantID := uuid.New()

	t.Run("successful creation", func(t *testing.T) {
		body := `{"amount":"10.50","currency":"USDT","provider":"TEST","merchantTradeNo":"ORDER-001"}`
		req := httptest.NewRequest(http.MethodPost, "/v1/payments", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+authToken(merchantID))
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusCreated, rec.Code)

		var resp map[string]any
		err := json.NewDecoder(rec.Body).Decode(&resp)
		require.NoError(t, err)

		data := resp["data"].(map[string]any)
		assert.NotEmpty(t, data["id"])
		assert.Equal(t, "10.5", data["amount"])
		assert.Equal(t, "INITIATED", data["status"])
		assert.NotEmpty(t, data["qrContent"])
		assert.NotEmpty(t, data["providerPayId"])
	})

	t.Run("missing auth", func(t *testing.T) {
		body := `{"amount":"10","currency":"USDT","provider":"TEST"}`
		req := httptest.NewRequest(http.MethodPost, "/v1/payments", bytes.NewBufferString(body))
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("invalid amount", func(t *testing.T) {
		body := `{"amount":"not-a-number","currency":"USDT","provider":"TEST"}`
		req := httptest.NewRequest(http.MethodPost, "/v1/payments", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+authToken(merchantID))
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}

func TestGetCheckoutHandler(t *testing.T) {
	svc := newStubService()
	h := handler.NewPaymentHandler(svc, svc.mockProv)
	router := handler.NewRouter(h, testJWTSecret)
	merchantID := uuid.New()

	// Create a payment first
	token := authToken(merchantID)
	body := `{"amount":"25","currency":"USDT","provider":"TEST"}`
	createReq := httptest.NewRequest(http.MethodPost, "/v1/payments", bytes.NewBufferString(body))
	createReq.Header.Set("Content-Type", "application/json")
	createReq.Header.Set("Authorization", "Bearer "+token)
	createRec := httptest.NewRecorder()
	router.ServeHTTP(createRec, createReq)
	require.Equal(t, http.StatusCreated, createRec.Code)

	var createResp map[string]any
	_ = json.NewDecoder(createRec.Body).Decode(&createResp)
	paymentID := createResp["data"].(map[string]any)["id"].(string)

	t.Run("get checkout data (no auth required)", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/payments/"+paymentID+"/checkout", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var resp map[string]any
		_ = json.NewDecoder(rec.Body).Decode(&resp)
		data := resp["data"].(map[string]any)
		assert.Equal(t, "INITIATED", data["status"])
		assert.NotEmpty(t, data["qrContent"])
	})

	t.Run("nonexistent payment", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/payments/"+uuid.New().String()+"/checkout", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusNotFound, rec.Code)
	})
}

func TestSimulatePaymentHandler(t *testing.T) {
	svc := newStubService()
	h := handler.NewPaymentHandler(svc, svc.mockProv)
	router := handler.NewRouter(h, testJWTSecret)
	merchantID := uuid.New()

	// Create payment
	token := authToken(merchantID)
	body := `{"amount":"50","currency":"USDT","provider":"TEST"}`
	createReq := httptest.NewRequest(http.MethodPost, "/v1/payments", bytes.NewBufferString(body))
	createReq.Header.Set("Content-Type", "application/json")
	createReq.Header.Set("Authorization", "Bearer "+token)
	createRec := httptest.NewRecorder()
	router.ServeHTTP(createRec, createReq)
	require.Equal(t, http.StatusCreated, createRec.Code)

	var createResp map[string]any
	_ = json.NewDecoder(createRec.Body).Decode(&createResp)
	paymentData := createResp["data"].(map[string]any)
	paymentID := paymentData["id"].(string)
	providerPayID := paymentData["providerPayId"].(string)

	t.Run("simulate then callback", func(t *testing.T) {
		// Simulate payment
		simReq := httptest.NewRequest(http.MethodPost, "/test/simulate/"+providerPayID, nil)
		simRec := httptest.NewRecorder()
		router.ServeHTTP(simRec, simReq)
		assert.Equal(t, http.StatusOK, simRec.Code)

		// Trigger callback
		cbReq := httptest.NewRequest(http.MethodPost, "/v1/payments/"+paymentID+"/callback", nil)
		cbRec := httptest.NewRecorder()
		router.ServeHTTP(cbRec, cbReq)
		assert.Equal(t, http.StatusOK, cbRec.Code)

		// Verify payment is now PAID
		getReq := httptest.NewRequest(http.MethodGet, "/v1/payments/"+paymentID+"/checkout", nil)
		getRec := httptest.NewRecorder()
		router.ServeHTTP(getRec, getReq)

		var getResp map[string]any
		_ = json.NewDecoder(getRec.Body).Decode(&getResp)
		data := getResp["data"].(map[string]any)
		assert.Equal(t, "PAID", data["status"])
	})
}
