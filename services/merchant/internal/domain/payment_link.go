package domain

import (
	"fmt"
	"net/url"
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
	MinAmount         *decimal.Decimal
	MaxAmount         *decimal.Decimal
	AllowQuantityBuy  bool
	MaxQuantity       int
	IsReusable        bool
	ShowOnQRPage      bool
	UsageCount        int
	Status            PaymentLinkStatus
	SuccessURL        string
	CancelURL         string
	WebhookURL        string
	MerchantTradeNo   string
	OrderExpireMinutes *int
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
		ID:          uuid.New(),
		MerchantID:  merchantID,
		Name:        name,
		Slug:        slug,
		Currency:    currency,
		Amount:      amount,
		MaxQuantity: 1,
		Status:      PaymentLinkActive,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

// Validate performs additional validation after all fields are set.
func (pl *PaymentLink) Validate() error {
	if pl.AllowCustomAmount {
		if pl.MinAmount != nil && pl.MinAmount.LessThanOrEqual(decimal.Zero) {
			return fmt.Errorf("%w: minAmount must be greater than zero", ErrInvalidPaymentLink)
		}
		if pl.MinAmount != nil && pl.MaxAmount != nil && pl.MaxAmount.LessThanOrEqual(*pl.MinAmount) {
			return fmt.Errorf("%w: maxAmount must be greater than minAmount", ErrInvalidPaymentLink)
		}
	}

	if pl.AllowQuantityBuy && pl.MaxQuantity < 1 {
		return fmt.Errorf("%w: maxQuantity must be at least 1 when allowQuantityBuy is true", ErrInvalidPaymentLink)
	}

	if pl.SuccessURL != "" {
		if _, err := url.ParseRequestURI(pl.SuccessURL); err != nil {
			return fmt.Errorf("%w: successUrl must be a valid URL", ErrInvalidPaymentLink)
		}
	}

	if pl.CancelURL != "" {
		if _, err := url.ParseRequestURI(pl.CancelURL); err != nil {
			return fmt.Errorf("%w: cancelUrl must be a valid URL", ErrInvalidPaymentLink)
		}
	}

	if pl.WebhookURL != "" {
		if _, err := url.ParseRequestURI(pl.WebhookURL); err != nil {
			return fmt.Errorf("%w: webhookUrl must be a valid URL", ErrInvalidPaymentLink)
		}
	}

	return nil
}

// IsExpired returns true if the link has passed its expiration date.
func (pl *PaymentLink) IsExpired() bool {
	return pl.ExpireAt != nil && time.Now().After(*pl.ExpireAt)
}

// IsConsumed returns true if this is a single-use link that has been used.
func (pl *PaymentLink) IsConsumed() bool {
	return !pl.IsReusable && pl.UsageCount > 0
}
