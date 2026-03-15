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

const testJWTSecret = "test-jwt-secret-key-at-least-32-chars"

// --- Stub service for testing handlers ---

type stubMerchantService struct {
	merchants map[uuid.UUID]*domain.Merchant
	users     map[uuid.UUID]*domain.User
}

func newStubService() *stubMerchantService {
	return &stubMerchantService{
		merchants: make(map[uuid.UUID]*domain.Merchant),
		users:     make(map[uuid.UUID]*domain.User),
	}
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

func (s *stubMerchantService) RegisterWithUser(_ context.Context, input service.RegisterWithUserInput) (*service.LoginResult, error) {
	m, err := domain.NewMerchant(input.BusinessName, input.ContactEmail)
	if err != nil {
		return nil, err
	}
	s.merchants[m.ID] = m

	u, err := domain.NewUser(m.ID, input.AdminEmail, input.AdminPassword, input.AdminName, domain.RoleAdmin, nil)
	if err != nil {
		return nil, err
	}
	s.users[u.ID] = u

	return &service.LoginResult{
		User:         u,
		Merchant:     m,
		AccessToken:  "test-access-token",
		RefreshToken: "test-refresh-token",
	}, nil
}

func (s *stubMerchantService) Login(_ context.Context, email, password string) (*service.LoginResult, error) {
	for _, u := range s.users {
		if u.Email == email && u.VerifyPassword(password) {
			m := s.merchants[u.MerchantID]
			return &service.LoginResult{
				User:         u,
				Merchant:     m,
				AccessToken:  "test-access-token",
				RefreshToken: "test-refresh-token",
			}, nil
		}
	}
	return nil, service.ErrInvalidCredentials
}

func (s *stubMerchantService) RefreshToken(_ context.Context, _ string) (*service.LoginResult, error) {
	return nil, service.ErrInvalidCredentials
}

func (s *stubMerchantService) GetByID(_ context.Context, id uuid.UUID) (*domain.Merchant, error) {
	m, ok := s.merchants[id]
	if !ok {
		return nil, domain.ErrMerchantNotFound
	}
	return m, nil
}

func (s *stubMerchantService) GetUserByID(_ context.Context, id uuid.UUID) (*domain.User, error) {
	u, ok := s.users[id]
	if !ok {
		return nil, domain.ErrUserNotFound
	}
	return u, nil
}

func (s *stubMerchantService) UpdateMerchantProfile(_ context.Context, id uuid.UUID, input service.UpdateProfileInput) (*domain.Merchant, error) {
	m, ok := s.merchants[id]
	if !ok {
		return nil, domain.ErrMerchantNotFound
	}
	if input.BusinessName != nil {
		m.BusinessName = *input.BusinessName
	}
	return m, nil
}

func TestRegisterWithUserHandler(t *testing.T) {
	svc := newStubService()
	h := handler.NewMerchantHandler(svc, testJWTSecret)
	router := handler.NewRouter(h, nil)

	t.Run("successful registration", func(t *testing.T) {
		body := `{"businessName":"Test Shop","email":"test@example.com","password":"SecurePass1","name":"John Doe","phone":"+94771234567"}`
		req := httptest.NewRequest(http.MethodPost, "/v1/auth/register", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusCreated, rec.Code)

		var resp map[string]any
		err := json.NewDecoder(rec.Body).Decode(&resp)
		require.NoError(t, err)

		data := resp["data"].(map[string]any)
		assert.NotEmpty(t, data["accessToken"])
		assert.NotEmpty(t, data["refreshToken"])

		merchant := data["merchant"].(map[string]any)
		assert.Equal(t, "Test Shop", merchant["businessName"])

		user := data["user"].(map[string]any)
		assert.Equal(t, "test@example.com", user["email"])
		assert.Equal(t, "ADMIN", user["role"])
	})

	t.Run("weak password", func(t *testing.T) {
		body := `{"businessName":"Shop","email":"weak@example.com","password":"weak","name":"John"}`
		req := httptest.NewRequest(http.MethodPost, "/v1/auth/register", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("malformed JSON", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/v1/auth/register", bytes.NewBufferString(`{invalid`))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}

func TestLoginHandler(t *testing.T) {
	svc := newStubService()
	h := handler.NewMerchantHandler(svc, testJWTSecret)
	router := handler.NewRouter(h, nil)

	// Register a user first
	regBody := `{"businessName":"Login Shop","email":"login@example.com","password":"SecurePass1","name":"John"}`
	regReq := httptest.NewRequest(http.MethodPost, "/v1/auth/register", bytes.NewBufferString(regBody))
	regReq.Header.Set("Content-Type", "application/json")
	regRec := httptest.NewRecorder()
	router.ServeHTTP(regRec, regReq)
	require.Equal(t, http.StatusCreated, regRec.Code)

	t.Run("successful login", func(t *testing.T) {
		body := `{"email":"login@example.com","password":"SecurePass1"}`
		req := httptest.NewRequest(http.MethodPost, "/v1/auth/login", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var resp map[string]any
		err := json.NewDecoder(rec.Body).Decode(&resp)
		require.NoError(t, err)
		data := resp["data"].(map[string]any)
		assert.NotEmpty(t, data["accessToken"])
	})

	t.Run("wrong password", func(t *testing.T) {
		body := `{"email":"login@example.com","password":"WrongPass1"}`
		req := httptest.NewRequest(http.MethodPost, "/v1/auth/login", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("nonexistent email", func(t *testing.T) {
		body := `{"email":"nobody@example.com","password":"Password1"}`
		req := httptest.NewRequest(http.MethodPost, "/v1/auth/login", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})
}
