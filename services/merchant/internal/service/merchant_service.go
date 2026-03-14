package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
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

// MerchantService orchestrates merchant domain operations.
type MerchantService struct {
	merchants MerchantRepository
	apiKeys   APIKeyRepository
	events    EventPublisher
}

// NewMerchantService creates a new MerchantService.
func NewMerchantService(merchants MerchantRepository, apiKeys APIKeyRepository, events EventPublisher) *MerchantService {
	return &MerchantService{
		merchants: merchants,
		apiKeys:   apiKeys,
		events:    events,
	}
}

// Register creates a new merchant.
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
// Returns the key entity and the plain-text secret (shown only once).
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
