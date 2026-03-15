package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/openlankapay/openlankapay/pkg/auth"
	"github.com/openlankapay/openlankapay/services/merchant/internal/domain"
)

// MerchantRepository defines the data access contract for merchants.
type MerchantRepository interface {
	Create(ctx context.Context, merchant *domain.Merchant) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Merchant, error)
	GetByEmail(ctx context.Context, email string) (*domain.Merchant, error)
	Update(ctx context.Context, merchant *domain.Merchant) error
	List(ctx context.Context, params ListParams) ([]*domain.Merchant, int, error)
	SoftDelete(ctx context.Context, id uuid.UUID) error
}

// APIKeyRepository defines the data access contract for API keys.
type APIKeyRepository interface {
	Create(ctx context.Context, key *domain.APIKey) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.APIKey, error)
	GetByKeyID(ctx context.Context, keyID string) (*domain.APIKey, error)
	ListByMerchant(ctx context.Context, merchantID uuid.UUID) ([]*domain.APIKey, error)
	Update(ctx context.Context, key *domain.APIKey) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// UserRepository defines the data access contract for users.
type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	GetByEmail(ctx context.Context, merchantID uuid.UUID, email string) (*domain.User, error)
	GetByEmailGlobal(ctx context.Context, email string) (*domain.User, error)
	ListByMerchant(ctx context.Context, merchantID uuid.UUID) ([]*domain.User, error)
	Update(ctx context.Context, user *domain.User) error
}

// EventPublisher defines the contract for publishing domain events.
type EventPublisher interface {
	Publish(ctx context.Context, subject string, data any) error
}

// ListParams holds pagination and filtering parameters.
type ListParams struct {
	Page      int
	PerPage   int
	KYCStatus *domain.KYCStatus
	Status    *domain.MerchantStatus
}

// RegisterInput holds the data required to register a new merchant.
type RegisterInput struct {
	BusinessName   string
	BusinessType   string
	RegistrationNo string
	Website        string
	ContactEmail   string
	ContactPhone   string
	ContactName    string
}

// RegisterWithUserInput holds data to register a merchant and create the admin user.
type RegisterWithUserInput struct {
	BusinessName string
	ContactEmail string
	ContactPhone string
	ContactName  string
	AdminEmail   string
	AdminPassword string
	AdminName    string
}

// LoginResult holds the result of a successful login.
type LoginResult struct {
	User         *domain.User
	Merchant     *domain.Merchant
	AccessToken  string
	RefreshToken string
}

// UpdateProfileInput holds data for updating merchant profile/KYC.
type UpdateProfileInput struct {
	BusinessName   *string
	BusinessType   *string
	RegistrationNo *string
	Website        *string
	ContactPhone   *string
	ContactName    *string
	AddressLine1   *string
	AddressLine2   *string
	City           *string
	District       *string
	PostalCode     *string
	BankName       *string
	BankBranch     *string
	BankAccountNo  *string
	BankAccountName *string
	SubmitKYC      bool
}

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrAccountInactive    = errors.New("account is inactive")
)

// MerchantService orchestrates merchant domain operations.
type MerchantService struct {
	merchants  MerchantRepository
	apiKeys    APIKeyRepository
	users      UserRepository
	events     EventPublisher
	jwtSecret  string
	tokenTTL   time.Duration
	refreshTTL time.Duration
}

// NewMerchantService creates a new MerchantService.
func NewMerchantService(
	merchants MerchantRepository,
	apiKeys APIKeyRepository,
	users UserRepository,
	events EventPublisher,
	jwtSecret string,
) *MerchantService {
	return &MerchantService{
		merchants:  merchants,
		apiKeys:    apiKeys,
		users:      users,
		events:     events,
		jwtSecret:  jwtSecret,
		tokenTTL:   24 * time.Hour,
		refreshTTL: 7 * 24 * time.Hour,
	}
}

// RegisterWithUser creates a new merchant and an admin user, returning JWT tokens.
func (s *MerchantService) RegisterWithUser(ctx context.Context, input RegisterWithUserInput) (*LoginResult, error) {
	merchant, err := domain.NewMerchant(input.BusinessName, input.ContactEmail)
	if err != nil {
		return nil, err
	}

	merchant.ContactPhone = input.ContactPhone
	merchant.ContactName = input.ContactName

	if err := s.merchants.Create(ctx, merchant); err != nil {
		return nil, err
	}

	user, err := domain.NewUser(merchant.ID, input.AdminEmail, input.AdminPassword, input.AdminName, domain.RoleAdmin, nil)
	if err != nil {
		return nil, err
	}

	if err := s.users.Create(ctx, user); err != nil {
		return nil, err
	}

	accessToken, err := auth.GenerateToken(user.ID, merchant.ID, user.Role, nil, s.jwtSecret, s.tokenTTL)
	if err != nil {
		return nil, fmt.Errorf("generating access token: %w", err)
	}

	refreshToken, err := auth.GenerateRefreshToken(user.ID, s.jwtSecret, s.refreshTTL)
	if err != nil {
		return nil, fmt.Errorf("generating refresh token: %w", err)
	}

	_ = s.events.Publish(ctx, "merchant.registered", merchant)

	return &LoginResult{
		User:         user,
		Merchant:     merchant,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// Login authenticates a user by email and password.
func (s *MerchantService) Login(ctx context.Context, email, password string) (*LoginResult, error) {
	user, err := s.users.GetByEmailGlobal(ctx, email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	if !user.IsActive {
		return nil, ErrAccountInactive
	}

	if !user.VerifyPassword(password) {
		return nil, ErrInvalidCredentials
	}

	merchant, err := s.merchants.GetByID(ctx, user.MerchantID)
	if err != nil {
		return nil, fmt.Errorf("fetching merchant: %w", err)
	}

	now := time.Now().UTC()
	user.LastLoginAt = &now
	_ = s.users.Update(ctx, user)

	accessToken, err := auth.GenerateToken(user.ID, merchant.ID, user.Role, user.BranchID, s.jwtSecret, s.tokenTTL)
	if err != nil {
		return nil, fmt.Errorf("generating access token: %w", err)
	}

	refreshToken, err := auth.GenerateRefreshToken(user.ID, s.jwtSecret, s.refreshTTL)
	if err != nil {
		return nil, fmt.Errorf("generating refresh token: %w", err)
	}

	return &LoginResult{
		User:         user,
		Merchant:     merchant,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// RefreshToken generates a new access token from a valid refresh token.
func (s *MerchantService) RefreshToken(ctx context.Context, refreshTokenStr string) (*LoginResult, error) {
	userID, err := auth.ValidateRefreshToken(refreshTokenStr, s.jwtSecret)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	if !user.IsActive {
		return nil, ErrAccountInactive
	}

	merchant, err := s.merchants.GetByID(ctx, user.MerchantID)
	if err != nil {
		return nil, fmt.Errorf("fetching merchant: %w", err)
	}

	accessToken, err := auth.GenerateToken(user.ID, merchant.ID, user.Role, user.BranchID, s.jwtSecret, s.tokenTTL)
	if err != nil {
		return nil, fmt.Errorf("generating access token: %w", err)
	}

	newRefreshToken, err := auth.GenerateRefreshToken(user.ID, s.jwtSecret, s.refreshTTL)
	if err != nil {
		return nil, fmt.Errorf("generating refresh token: %w", err)
	}

	return &LoginResult{
		User:         user,
		Merchant:     merchant,
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
	}, nil
}

// GetUserByID returns a user by ID.
func (s *MerchantService) GetUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	return s.users.GetByID(ctx, id)
}

// UpdateMerchantProfile updates merchant profile and optionally submits KYC.
func (s *MerchantService) UpdateMerchantProfile(ctx context.Context, id uuid.UUID, input UpdateProfileInput) (*domain.Merchant, error) {
	merchant, err := s.merchants.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if input.BusinessName != nil {
		merchant.BusinessName = *input.BusinessName
	}
	if input.BusinessType != nil {
		merchant.BusinessType = *input.BusinessType
	}
	if input.RegistrationNo != nil {
		merchant.RegistrationNo = *input.RegistrationNo
	}
	if input.Website != nil {
		merchant.Website = *input.Website
	}
	if input.ContactPhone != nil {
		merchant.ContactPhone = *input.ContactPhone
	}
	if input.ContactName != nil {
		merchant.ContactName = *input.ContactName
	}
	if input.AddressLine1 != nil {
		merchant.AddressLine1 = *input.AddressLine1
	}
	if input.AddressLine2 != nil {
		merchant.AddressLine2 = *input.AddressLine2
	}
	if input.City != nil {
		merchant.City = *input.City
	}
	if input.District != nil {
		merchant.District = *input.District
	}
	if input.PostalCode != nil {
		merchant.PostalCode = *input.PostalCode
	}
	if input.BankName != nil {
		merchant.BankName = *input.BankName
	}
	if input.BankBranch != nil {
		merchant.BankBranch = *input.BankBranch
	}
	if input.BankAccountNo != nil {
		merchant.BankAccountNo = *input.BankAccountNo
	}
	if input.BankAccountName != nil {
		merchant.BankAccountName = *input.BankAccountName
	}

	if input.SubmitKYC && merchant.KYCStatus == domain.KYCPending {
		now := time.Now().UTC()
		merchant.KYCSubmittedAt = &now
		if err := merchant.TransitionKYC(domain.KYCInstantAccess); err != nil {
			return nil, err
		}
	}

	if err := s.merchants.Update(ctx, merchant); err != nil {
		return nil, err
	}

	return merchant, nil
}

// Register creates a new merchant (legacy, without user).
func (s *MerchantService) Register(ctx context.Context, input RegisterInput) (*domain.Merchant, error) {
	merchant, err := domain.NewMerchant(input.BusinessName, input.ContactEmail)
	if err != nil {
		return nil, err
	}

	merchant.BusinessType = input.BusinessType
	merchant.RegistrationNo = input.RegistrationNo
	merchant.Website = input.Website
	merchant.ContactPhone = input.ContactPhone
	merchant.ContactName = input.ContactName

	if err := s.merchants.Create(ctx, merchant); err != nil {
		return nil, err
	}

	_ = s.events.Publish(ctx, "merchant.registered", merchant)

	return merchant, nil
}

// GetByID returns a merchant by ID.
func (s *MerchantService) GetByID(ctx context.Context, id uuid.UUID) (*domain.Merchant, error) {
	return s.merchants.GetByID(ctx, id)
}

// List returns paginated merchants.
func (s *MerchantService) List(ctx context.Context, params ListParams) ([]*domain.Merchant, int, error) {
	return s.merchants.List(ctx, params)
}

// Approve transitions a merchant's KYC status to APPROVED.
func (s *MerchantService) Approve(ctx context.Context, id uuid.UUID) error {
	merchant, err := s.merchants.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if err := merchant.TransitionKYC(domain.KYCApproved); err != nil {
		return err
	}

	if err := s.merchants.Update(ctx, merchant); err != nil {
		return err
	}

	_ = s.events.Publish(ctx, "merchant.approved", merchant)

	return nil
}

// Reject transitions a merchant's KYC status to REJECTED with a reason.
func (s *MerchantService) Reject(ctx context.Context, id uuid.UUID, reason string) error {
	merchant, err := s.merchants.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if err := merchant.TransitionKYC(domain.KYCRejected); err != nil {
		return err
	}

	merchant.KYCRejectionReason = reason

	if err := s.merchants.Update(ctx, merchant); err != nil {
		return err
	}

	_ = s.events.Publish(ctx, "merchant.rejected", merchant)

	return nil
}

// CreateAPIKey generates a new API key for a merchant.
func (s *MerchantService) CreateAPIKey(ctx context.Context, merchantID uuid.UUID, env, name string) (*domain.APIKey, string, error) {
	_, err := s.merchants.GetByID(ctx, merchantID)
	if err != nil {
		return nil, "", err
	}

	key, plainSecret, err := domain.NewAPIKey(merchantID, env, name)
	if err != nil {
		return nil, "", fmt.Errorf("generating API key: %w", err)
	}

	if err := s.apiKeys.Create(ctx, key); err != nil {
		return nil, "", fmt.Errorf("storing API key: %w", err)
	}

	_ = s.events.Publish(ctx, "merchant.apikey.created", key)

	return key, plainSecret, nil
}

// ValidateAPIKey checks a key ID and secret, returning the associated merchant.
func (s *MerchantService) ValidateAPIKey(ctx context.Context, keyID, secret string) (*domain.Merchant, error) {
	key, err := s.apiKeys.GetByKeyID(ctx, keyID)
	if err != nil {
		return nil, fmt.Errorf("invalid API key: %w", err)
	}

	if !key.IsActive {
		return nil, errors.New("API key is revoked")
	}

	if !key.VerifySecret(secret) {
		return nil, errors.New("invalid API secret")
	}

	merchant, err := s.merchants.GetByID(ctx, key.MerchantID)
	if err != nil {
		return nil, fmt.Errorf("merchant not found for key: %w", err)
	}

	return merchant, nil
}

// RevokeAPIKey deactivates an API key.
func (s *MerchantService) RevokeAPIKey(ctx context.Context, id uuid.UUID, reason string) error {
	key, err := s.apiKeys.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if err := key.Revoke(reason); err != nil {
		return err
	}

	return s.apiKeys.Update(ctx, key)
}
