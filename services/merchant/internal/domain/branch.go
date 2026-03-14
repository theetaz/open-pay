package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Branch represents a merchant's business location.
type Branch struct {
	ID              uuid.UUID
	MerchantID      uuid.UUID
	Name            string
	Description     string
	Address         string
	City            string
	BankName        string
	BankBranch      string
	BankAccountNo   string
	BankAccountName string
	IsActive        bool
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       *time.Time
}

// NewBranch creates a validated Branch entity.
func NewBranch(merchantID uuid.UUID, name string) (*Branch, error) {
	if merchantID == uuid.Nil {
		return nil, fmt.Errorf("%w: merchant ID is required", ErrInvalidBranch)
	}
	if name == "" {
		return nil, fmt.Errorf("%w: name is required", ErrInvalidBranch)
	}

	now := time.Now().UTC()
	return &Branch{
		ID:         uuid.New(),
		MerchantID: merchantID,
		Name:       name,
		IsActive:   true,
		CreatedAt:  now,
		UpdatedAt:  now,
	}, nil
}
