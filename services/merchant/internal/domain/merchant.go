package domain

import (
	"fmt"
	"net/mail"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// KYC status values.
type KYCStatus string

const (
	KYCPending      KYCStatus = "PENDING"
	KYCUnderReview  KYCStatus = "UNDER_REVIEW"
	KYCApproved     KYCStatus = "APPROVED"
	KYCRejected     KYCStatus = "REJECTED"
	KYCInstantAccess KYCStatus = "INSTANT_ACCESS"
)

// Merchant status values.
type MerchantStatus string

const (
	MerchantActive   MerchantStatus = "ACTIVE"
	MerchantInactive MerchantStatus = "INACTIVE"
)

// Valid KYC transitions.
var validKYCTransitions = map[KYCStatus][]KYCStatus{
	KYCPending:       {KYCUnderReview, KYCInstantAccess},
	KYCUnderReview:   {KYCApproved, KYCRejected},
	KYCInstantAccess: {KYCUnderReview, KYCApproved},
}

// Merchant represents a registered business.
type Merchant struct {
	ID               uuid.UUID
	BusinessName     string
	BusinessType     string
	RegistrationNo   string
	Website          string
	ContactEmail     string
	ContactPhone     string
	ContactName      string
	AddressLine1     string
	AddressLine2     string
	City             string
	District         string
	PostalCode       string
	Country          string
	KYCStatus        KYCStatus
	KYCSubmittedAt   *time.Time
	KYCReviewedAt    *time.Time
	KYCRejectionReason string
	TransactionLimitUSDT *decimal.Decimal
	DailyLimitUSDT       *decimal.Decimal
	MonthlyLimitUSDT     *decimal.Decimal
	InstantAccessRemaining decimal.Decimal
	BankName          string
	BankBranch        string
	BankAccountNo     string
	BankAccountName   string
	DefaultCurrency   string
	DefaultProvider   string
	Status            MerchantStatus
	CreatedAt         time.Time
	UpdatedAt         time.Time
	DeletedAt         *time.Time
}

// NewMerchant creates a validated Merchant entity.
func NewMerchant(businessName, contactEmail string) (*Merchant, error) {
	if businessName == "" {
		return nil, fmt.Errorf("%w: business name is required", ErrInvalidMerchant)
	}

	if err := validateEmail(contactEmail); err != nil {
		return nil, fmt.Errorf("%w: %s", ErrInvalidMerchant, err.Error())
	}

	now := time.Now().UTC()
	return &Merchant{
		ID:                     uuid.New(),
		BusinessName:           businessName,
		ContactEmail:           contactEmail,
		Country:                "LK",
		KYCStatus:              KYCPending,
		InstantAccessRemaining: decimal.NewFromInt(5000),
		DefaultCurrency:        "LKR",
		Status:                 MerchantActive,
		CreatedAt:              now,
		UpdatedAt:              now,
	}, nil
}

// TransitionKYC moves the merchant to a new KYC status if the transition is valid.
func (m *Merchant) TransitionKYC(to KYCStatus) error {
	allowed, ok := validKYCTransitions[m.KYCStatus]
	if !ok {
		return fmt.Errorf("%w: no transitions from %s", ErrInvalidKYCTransition, m.KYCStatus)
	}

	for _, s := range allowed {
		if s == to {
			m.KYCStatus = to
			now := time.Now().UTC()
			m.UpdatedAt = now
			if to == KYCApproved || to == KYCRejected {
				m.KYCReviewedAt = &now
			}
			return nil
		}
	}

	return fmt.Errorf("%w: cannot transition from %s to %s", ErrInvalidKYCTransition, m.KYCStatus, to)
}

// CanAcceptPayments checks if the merchant is eligible to process payments.
func (m *Merchant) CanAcceptPayments() bool {
	if m.Status != MerchantActive {
		return false
	}
	return m.KYCStatus == KYCApproved || m.KYCStatus == KYCInstantAccess
}

func validateEmail(email string) error {
	if email == "" {
		return fmt.Errorf("email is required")
	}
	_, err := mail.ParseAddress(email)
	if err != nil {
		return fmt.Errorf("invalid email format")
	}
	return nil
}
