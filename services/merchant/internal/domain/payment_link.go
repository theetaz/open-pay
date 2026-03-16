package domain

import (
	"fmt"
	"regexp"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// PaymentLinkStatus values.
type PaymentLinkStatus string

const (
	PaymentLinkActive   PaymentLinkStatus = "ACTIVE"
	PaymentLinkInactive PaymentLinkStatus = "INACTIVE"
)

// PaymentLink represents a shareable payment link created by a merchant.
type PaymentLink struct {
	ID                uuid.UUID
	MerchantID        uuid.UUID
	BranchID          *uuid.UUID
	Name              string
	Slug              string
	Description       string
	Currency          string
	Amount            decimal.Decimal
	AllowCustomAmount bool
	IsReusable        bool
	ShowOnQRPage      bool
	UsageCount        int
	Status            PaymentLinkStatus
	ExpireAt          *time.Time
	CreatedAt         time.Time
	UpdatedAt         time.Time
	DeletedAt         *time.Time
}

var slugRegex = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

// NewPaymentLink creates a validated PaymentLink.
func NewPaymentLink(merchantID uuid.UUID, name, slug, currency string, amount decimal.Decimal) (*PaymentLink, error) {
	if name == "" {
		return nil, fmt.Errorf("%w: name is required", ErrInvalidPaymentLink)
	}
	if slug == "" {
		return nil, fmt.Errorf("%w: slug is required", ErrInvalidPaymentLink)
	}
	if !slugRegex.MatchString(slug) {
		return nil, fmt.Errorf("%w: slug must be lowercase alphanumeric with hyphens", ErrInvalidPaymentLink)
	}
	if len(slug) < 2 || len(slug) > 100 {
		return nil, fmt.Errorf("%w: slug must be between 2 and 100 characters", ErrInvalidPaymentLink)
	}
	if currency != "LKR" && currency != "USDT" {
		return nil, fmt.Errorf("%w: currency must be LKR or USDT", ErrInvalidPaymentLink)
	}
	if amount.LessThanOrEqual(decimal.Zero) {
		return nil, fmt.Errorf("%w: amount must be greater than zero", ErrInvalidPaymentLink)
	}

	now := time.Now().UTC()
	return &PaymentLink{
		ID:         uuid.New(),
		MerchantID: merchantID,
		Name:       name,
		Slug:       slug,
		Currency:   currency,
		Amount:     amount,
		Status:     PaymentLinkActive,
		CreatedAt:  now,
		UpdatedAt:  now,
	}, nil
}
