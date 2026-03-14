package port

import (
	"context"

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
	GetByKeyID(ctx context.Context, keyID string) (*domain.APIKey, error)
	ListByMerchant(ctx context.Context, merchantID uuid.UUID) ([]*domain.APIKey, error)
	Update(ctx context.Context, key *domain.APIKey) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// BranchRepository defines the data access contract for branches.
type BranchRepository interface {
	Create(ctx context.Context, branch *domain.Branch) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Branch, error)
	ListByMerchant(ctx context.Context, merchantID uuid.UUID) ([]*domain.Branch, error)
	Update(ctx context.Context, branch *domain.Branch) error
	SoftDelete(ctx context.Context, id uuid.UUID) error
}

// UserRepository defines the data access contract for dashboard users.
type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	GetByEmail(ctx context.Context, merchantID uuid.UUID, email string) (*domain.User, error)
	ListByMerchant(ctx context.Context, merchantID uuid.UUID) ([]*domain.User, error)
	Update(ctx context.Context, user *domain.User) error
}

// ListParams holds pagination and filtering parameters.
type ListParams struct {
	Page      int
	PerPage   int
	KYCStatus *domain.KYCStatus
	Status    *domain.MerchantStatus
}
