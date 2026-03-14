package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/openlankapay/openlankapay/services/merchant/internal/domain"
	"github.com/openlankapay/openlankapay/services/merchant/internal/handler"
	"github.com/openlankapay/openlankapay/services/merchant/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Stub service for testing handlers ---

type stubMerchantService struct {
	merchants map[uuid.UUID]*domain.Merchant
}

func newStubService() *stubMerchantService {
	return &stubMerchantService{merchants: make(map[uuid.UUID]*domain.Merchant)}
}

func (s *stubMerchantService) Register(_ context.Context, input service.RegisterInput) (*domain.Merchant, error) {
	m, err := domain.NewMerchant(input.BusinessName, input.ContactEmail)
	if err != nil {
		return nil, err
	}
	m.ContactPhone = input.ContactPhone
	m.ContactName = input.ContactName
	s.merchants[m.ID] = m
	return m, nil
}

func (s *stubMerchantService) GetByID(_ context.Context, id uuid.UUID) (*domain.Merchant, error) {
	m, ok := s.merchants[id]
	if !ok {
		return nil, domain.ErrMerchantNotFound
	}
	return m, nil
}

func TestRegisterHandler(t *testing.T) {
	svc := newStubService()
	h := handler.NewMerchantHandler(svc)
	router := handler.NewRouter(h)

	t.Run("successful registration", func(t *testing.T) {
		body := `{"businessName":"Test Shop","contactEmail":"test@example.com","contactPhone":"+94771234567","contactName":"John"}`
		req := httptest.NewRequest(http.MethodPost, "/v1/merchants", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusCreated, rec.Code)

		var resp map[string]any
		err := json.NewDecoder(rec.Body).Decode(&resp)
		require.NoError(t, err)

		data := resp["data"].(map[string]any)
		assert.Equal(t, "Test Shop", data["businessName"])
		assert.Equal(t, "PENDING", data["kycStatus"])
		assert.NotEmpty(t, data["id"])
	})

	t.Run("invalid request body", func(t *testing.T) {
		body := `{"businessName":"","contactEmail":"bad"}`
		req := httptest.NewRequest(http.MethodPost, "/v1/merchants", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("malformed JSON", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/v1/merchants", bytes.NewBufferString(`{invalid`))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}

func TestGetMerchantHandler(t *testing.T) {
	svc := newStubService()
	h := handler.NewMerchantHandler(svc)
	router := handler.NewRouter(h)

	// Register a merchant first
	m, _ := svc.Register(context.Background(), service.RegisterInput{
		BusinessName: "Test Shop",
		ContactEmail: "get@example.com",
	})

	t.Run("get existing merchant", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/merchants/"+m.ID.String(), nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var resp map[string]any
		err := json.NewDecoder(rec.Body).Decode(&resp)
		require.NoError(t, err)
		data := resp["data"].(map[string]any)
		assert.Equal(t, "Test Shop", data["businessName"])
	})

	t.Run("get nonexistent merchant", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/merchants/"+uuid.New().String(), nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("invalid UUID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/merchants/not-a-uuid", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}
