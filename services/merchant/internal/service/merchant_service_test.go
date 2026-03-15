package service_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/openlankapay/openlankapay/services/merchant/internal/domain"
	"github.com/openlankapay/openlankapay/services/merchant/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testJWTSecret = "test-jwt-secret-key-at-least-32-chars"

// --- Mock implementations ---

type mockMerchantRepo struct {
	merchants map[uuid.UUID]*domain.Merchant
	byEmail   map[string]*domain.Merchant
}

func newMockMerchantRepo() *mockMerchantRepo {
	return &mockMerchantRepo{
		merchants: make(map[uuid.UUID]*domain.Merchant),
		byEmail:   make(map[string]*domain.Merchant),
	}
}

func (m *mockMerchantRepo) Create(_ context.Context, merchant *domain.Merchant) error {
	if _, exists := m.byEmail[merchant.ContactEmail]; exists {
		return domain.ErrDuplicateEmail
	}
	m.merchants[merchant.ID] = merchant
	m.byEmail[merchant.ContactEmail] = merchant
	return nil
}

func (m *mockMerchantRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.Merchant, error) {
	merchant, ok := m.merchants[id]
	if !ok {
		return nil, domain.ErrMerchantNotFound
	}
	return merchant, nil
}

func (m *mockMerchantRepo) GetByEmail(_ context.Context, email string) (*domain.Merchant, error) {
	merchant, ok := m.byEmail[email]
	if !ok {
		return nil, domain.ErrMerchantNotFound
	}
	return merchant, nil
}

func (m *mockMerchantRepo) Update(_ context.Context, merchant *domain.Merchant) error {
	m.merchants[merchant.ID] = merchant
	return nil
}

func (m *mockMerchantRepo) List(_ context.Context, _ service.ListParams) ([]*domain.Merchant, int, error) {
	var result []*domain.Merchant
	for _, v := range m.merchants {
		result = append(result, v)
	}
	return result, len(result), nil
}

func (m *mockMerchantRepo) SoftDelete(_ context.Context, _ uuid.UUID) error {
	return nil
}

type mockAPIKeyRepo struct {
	keys    map[uuid.UUID]*domain.APIKey
	byKeyID map[string]*domain.APIKey
}

func newMockAPIKeyRepo() *mockAPIKeyRepo {
	return &mockAPIKeyRepo{
		keys:    make(map[uuid.UUID]*domain.APIKey),
		byKeyID: make(map[string]*domain.APIKey),
	}
}

func (m *mockAPIKeyRepo) Create(_ context.Context, key *domain.APIKey) error {
	m.keys[key.ID] = key
	m.byKeyID[key.KeyID] = key
	return nil
}

func (m *mockAPIKeyRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.APIKey, error) {
	key, ok := m.keys[id]
	if !ok {
		return nil, domain.ErrAPIKeyNotFound
	}
	return key, nil
}

func (m *mockAPIKeyRepo) GetByKeyID(_ context.Context, keyID string) (*domain.APIKey, error) {
	key, ok := m.byKeyID[keyID]
	if !ok {
		return nil, domain.ErrAPIKeyNotFound
	}
	return key, nil
}

func (m *mockAPIKeyRepo) ListByMerchant(_ context.Context, merchantID uuid.UUID) ([]*domain.APIKey, error) {
	var result []*domain.APIKey
	for _, k := range m.keys {
		if k.MerchantID == merchantID {
			result = append(result, k)
		}
	}
	return result, nil
}

func (m *mockAPIKeyRepo) Update(_ context.Context, key *domain.APIKey) error {
	m.keys[key.ID] = key
	m.byKeyID[key.KeyID] = key
	return nil
}

func (m *mockAPIKeyRepo) Delete(_ context.Context, id uuid.UUID) error {
	key, ok := m.keys[id]
	if ok {
		delete(m.byKeyID, key.KeyID)
		delete(m.keys, id)
	}
	return nil
}

type mockUserRepo struct {
	users   map[uuid.UUID]*domain.User
	byEmail map[string]*domain.User
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{
		users:   make(map[uuid.UUID]*domain.User),
		byEmail: make(map[string]*domain.User),
	}
}

func (m *mockUserRepo) Create(_ context.Context, user *domain.User) error {
	key := user.MerchantID.String() + ":" + user.Email
	if _, exists := m.byEmail[key]; exists {
		return domain.ErrDuplicateEmail
	}
	m.users[user.ID] = user
	m.byEmail[key] = user
	m.byEmail["global:"+user.Email] = user
	return nil
}

func (m *mockUserRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.User, error) {
	user, ok := m.users[id]
	if !ok {
		return nil, domain.ErrUserNotFound
	}
	return user, nil
}

func (m *mockUserRepo) GetByEmail(_ context.Context, merchantID uuid.UUID, email string) (*domain.User, error) {
	key := merchantID.String() + ":" + email
	user, ok := m.byEmail[key]
	if !ok {
		return nil, domain.ErrUserNotFound
	}
	return user, nil
}

func (m *mockUserRepo) GetByEmailGlobal(_ context.Context, email string) (*domain.User, error) {
	user, ok := m.byEmail["global:"+email]
	if !ok {
		return nil, domain.ErrUserNotFound
	}
	return user, nil
}

func (m *mockUserRepo) ListByMerchant(_ context.Context, merchantID uuid.UUID) ([]*domain.User, error) {
	var result []*domain.User
	for _, u := range m.users {
		if u.MerchantID == merchantID {
			result = append(result, u)
		}
	}
	return result, nil
}

func (m *mockUserRepo) Update(_ context.Context, user *domain.User) error {
	m.users[user.ID] = user
	return nil
}

type mockEventPublisher struct {
	published []publishedEvent
}

type publishedEvent struct {
	subject string
	data    any
}

func (m *mockEventPublisher) Publish(_ context.Context, subject string, data any) error {
	m.published = append(m.published, publishedEvent{subject, data})
	return nil
}

func newTestService() *service.MerchantService {
	return service.NewMerchantService(
		newMockMerchantRepo(),
		newMockAPIKeyRepo(),
		newMockUserRepo(),
		&mockEventPublisher{},
		testJWTSecret,
	)
}

// --- Tests ---

func TestRegisterMerchant(t *testing.T) {
	ctx := context.Background()

	t.Run("successful registration", func(t *testing.T) {
		svc := newTestService()

		merchant, err := svc.Register(ctx, service.RegisterInput{
			BusinessName: "Test Shop",
			ContactEmail: "shop@example.com",
			ContactPhone: "+94771234567",
			ContactName:  "John Doe",
		})
		require.NoError(t, err)
		assert.Equal(t, "Test Shop", merchant.BusinessName)
		assert.Equal(t, domain.KYCPending, merchant.KYCStatus)
		assert.NotEmpty(t, merchant.ID)
	})

	t.Run("duplicate email", func(t *testing.T) {
		svc := service.NewMerchantService(newMockMerchantRepo(), newMockAPIKeyRepo(), newMockUserRepo(), &mockEventPublisher{}, testJWTSecret)

		_, err := svc.Register(ctx, service.RegisterInput{
			BusinessName: "Shop 1",
			ContactEmail: "dup@example.com",
		})
		require.NoError(t, err)

		_, err = svc.Register(ctx, service.RegisterInput{
			BusinessName: "Shop 2",
			ContactEmail: "dup@example.com",
		})
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrDuplicateEmail)
	})

	t.Run("publishes event", func(t *testing.T) {
		pub := &mockEventPublisher{}
		svc := service.NewMerchantService(newMockMerchantRepo(), newMockAPIKeyRepo(), newMockUserRepo(), pub, testJWTSecret)

		_, err := svc.Register(ctx, service.RegisterInput{
			BusinessName: "Test Shop",
			ContactEmail: "events@example.com",
		})
		require.NoError(t, err)
		require.Len(t, pub.published, 1)
		assert.Equal(t, "merchant.registered", pub.published[0].subject)
	})
}

func TestRegisterWithUser(t *testing.T) {
	ctx := context.Background()

	t.Run("successful registration with user", func(t *testing.T) {
		svc := newTestService()

		result, err := svc.RegisterWithUser(ctx, service.RegisterWithUserInput{
			BusinessName: "New Shop",
			ContactEmail: "newshop@example.com",
			AdminEmail:   "admin@newshop.com",
			AdminPassword: "SecurePass1",
			AdminName:    "Admin User",
		})
		require.NoError(t, err)
		assert.Equal(t, "New Shop", result.Merchant.BusinessName)
		assert.Equal(t, "admin@newshop.com", result.User.Email)
		assert.Equal(t, domain.RoleAdmin, result.User.Role)
		assert.NotEmpty(t, result.AccessToken)
		assert.NotEmpty(t, result.RefreshToken)
	})

	t.Run("weak password fails", func(t *testing.T) {
		svc := newTestService()

		_, err := svc.RegisterWithUser(ctx, service.RegisterWithUserInput{
			BusinessName: "Shop",
			ContactEmail: "shop2@example.com",
			AdminEmail:   "admin@shop2.com",
			AdminPassword: "weak",
			AdminName:    "Admin",
		})
		require.Error(t, err)
	})
}

func TestLogin(t *testing.T) {
	ctx := context.Background()

	t.Run("successful login", func(t *testing.T) {
		svc := newTestService()

		_, err := svc.RegisterWithUser(ctx, service.RegisterWithUserInput{
			BusinessName: "Login Shop",
			ContactEmail: "login@example.com",
			AdminEmail:   "admin@login.com",
			AdminPassword: "SecurePass1",
			AdminName:    "Admin",
		})
		require.NoError(t, err)

		result, err := svc.Login(ctx, "admin@login.com", "SecurePass1")
		require.NoError(t, err)
		assert.Equal(t, "admin@login.com", result.User.Email)
		assert.NotEmpty(t, result.AccessToken)
		assert.NotNil(t, result.User.LastLoginAt)
	})

	t.Run("wrong password", func(t *testing.T) {
		svc := newTestService()

		_, _ = svc.RegisterWithUser(ctx, service.RegisterWithUserInput{
			BusinessName: "Wrong Pass Shop",
			ContactEmail: "wp@example.com",
			AdminEmail:   "admin@wp.com",
			AdminPassword: "SecurePass1",
			AdminName:    "Admin",
		})

		_, err := svc.Login(ctx, "admin@wp.com", "WrongPass1")
		require.Error(t, err)
		assert.ErrorIs(t, err, service.ErrInvalidCredentials)
	})

	t.Run("nonexistent email", func(t *testing.T) {
		svc := newTestService()

		_, err := svc.Login(ctx, "nobody@example.com", "Password1")
		require.Error(t, err)
		assert.ErrorIs(t, err, service.ErrInvalidCredentials)
	})
}

func TestRefreshToken(t *testing.T) {
	ctx := context.Background()

	t.Run("valid refresh token", func(t *testing.T) {
		svc := newTestService()

		reg, _ := svc.RegisterWithUser(ctx, service.RegisterWithUserInput{
			BusinessName: "Refresh Shop",
			ContactEmail: "refresh@example.com",
			AdminEmail:   "admin@refresh.com",
			AdminPassword: "SecurePass1",
			AdminName:    "Admin",
		})

		result, err := svc.RefreshToken(ctx, reg.RefreshToken)
		require.NoError(t, err)
		assert.NotEmpty(t, result.AccessToken)
		assert.NotEmpty(t, result.RefreshToken)
	})

	t.Run("invalid refresh token", func(t *testing.T) {
		svc := newTestService()

		_, err := svc.RefreshToken(ctx, "invalid-token")
		require.Error(t, err)
	})
}

func TestApproveMerchant(t *testing.T) {
	ctx := context.Background()

	t.Run("approve under review merchant", func(t *testing.T) {
		repo := newMockMerchantRepo()
		pub := &mockEventPublisher{}
		svc := service.NewMerchantService(repo, newMockAPIKeyRepo(), newMockUserRepo(), pub, testJWTSecret)

		merchant, _ := svc.Register(ctx, service.RegisterInput{
			BusinessName: "Test Shop",
			ContactEmail: "approve@example.com",
		})

		// Move to under_review first
		merchant.KYCStatus = domain.KYCUnderReview
		_ = repo.Update(ctx, merchant)

		err := svc.Approve(ctx, merchant.ID)
		require.NoError(t, err)

		updated, _ := repo.GetByID(ctx, merchant.ID)
		assert.Equal(t, domain.KYCApproved, updated.KYCStatus)
	})

	t.Run("cannot approve pending merchant", func(t *testing.T) {
		svc := newTestService()

		merchant, _ := svc.Register(ctx, service.RegisterInput{
			BusinessName: "Test Shop",
			ContactEmail: "pending@example.com",
		})

		err := svc.Approve(ctx, merchant.ID)
		require.Error(t, err)
	})
}

func TestCreateAPIKey(t *testing.T) {
	ctx := context.Background()

	t.Run("create live key", func(t *testing.T) {
		svc := newTestService()

		merchant, _ := svc.Register(ctx, service.RegisterInput{
			BusinessName: "Test Shop",
			ContactEmail: "keys@example.com",
		})

		key, secret, err := svc.CreateAPIKey(ctx, merchant.ID, "live", "Production")
		require.NoError(t, err)
		assert.Contains(t, key.KeyID, "ak_live_")
		assert.Contains(t, secret, "sk_live_")
	})

	t.Run("create key for nonexistent merchant", func(t *testing.T) {
		svc := newTestService()

		_, _, err := svc.CreateAPIKey(ctx, uuid.New(), "live", "Key")
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrMerchantNotFound)
	})
}

func TestValidateAPIKey(t *testing.T) {
	ctx := context.Background()

	t.Run("valid key returns merchant", func(t *testing.T) {
		svc := newTestService()

		merchant, _ := svc.Register(ctx, service.RegisterInput{
			BusinessName: "Test Shop",
			ContactEmail: "validate@example.com",
		})

		key, secret, _ := svc.CreateAPIKey(ctx, merchant.ID, "live", "Key")

		result, err := svc.ValidateAPIKey(ctx, key.KeyID, secret)
		require.NoError(t, err)
		assert.Equal(t, merchant.ID, result.ID)
	})

	t.Run("wrong secret fails", func(t *testing.T) {
		svc := newTestService()

		merchant, _ := svc.Register(ctx, service.RegisterInput{
			BusinessName: "Test Shop",
			ContactEmail: "wrong@example.com",
		})

		key, _, _ := svc.CreateAPIKey(ctx, merchant.ID, "live", "Key")

		_, err := svc.ValidateAPIKey(ctx, key.KeyID, "wrong_secret")
		require.Error(t, err)
	})

	t.Run("revoked key fails", func(t *testing.T) {
		svc := newTestService()

		merchant, _ := svc.Register(ctx, service.RegisterInput{
			BusinessName: "Test Shop",
			ContactEmail: "revoked@example.com",
		})

		key, secret, _ := svc.CreateAPIKey(ctx, merchant.ID, "live", "Key")
		_ = svc.RevokeAPIKey(ctx, key.ID, "Test")

		_, err := svc.ValidateAPIKey(ctx, key.KeyID, secret)
		require.Error(t, err)
	})
}
