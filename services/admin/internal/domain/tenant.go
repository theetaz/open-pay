package domain

import (
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidTenant  = errors.New("invalid tenant data")
	ErrTenantNotFound = errors.New("tenant not found")
	ErrDuplicateSlug  = errors.New("tenant slug already in use")
)

// slugPattern allows lowercase alphanumeric characters and hyphens.
var slugPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

// hexColorPattern matches a 7-character hex color code (e.g. #6366F1).
var hexColorPattern = regexp.MustCompile(`^#[0-9A-Fa-f]{6}$`)

// Tenant represents a white-label tenant on the platform.
type Tenant struct {
	ID             uuid.UUID
	Name           string
	Slug           string
	Domain         *string
	LogoURL        *string
	PrimaryColor   string
	PlatformFeePct float64
	ExchangeFeePct float64
	IsActive       bool
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// NewTenantInput holds the data required to create a new tenant.
type NewTenantInput struct {
	Name           string
	Slug           string
	Domain         *string
	LogoURL        *string
	PrimaryColor   string
	PlatformFeePct float64
	ExchangeFeePct float64
}

// NewTenant creates a validated Tenant from the given input.
func NewTenant(input NewTenantInput) (*Tenant, error) {
	if input.Name == "" {
		return nil, fmt.Errorf("%w: name is required", ErrInvalidTenant)
	}
	if input.Slug == "" {
		return nil, fmt.Errorf("%w: slug is required", ErrInvalidTenant)
	}
	if len(input.Slug) > 100 {
		return nil, fmt.Errorf("%w: slug must be 100 characters or fewer", ErrInvalidTenant)
	}
	if !slugPattern.MatchString(input.Slug) {
		return nil, fmt.Errorf("%w: slug must contain only lowercase alphanumeric characters and hyphens", ErrInvalidTenant)
	}

	if input.PrimaryColor == "" {
		input.PrimaryColor = "#6366F1"
	}
	if !hexColorPattern.MatchString(input.PrimaryColor) {
		return nil, fmt.Errorf("%w: primary_color must be a valid hex color (e.g. #6366F1)", ErrInvalidTenant)
	}

	if input.PlatformFeePct < 0 {
		return nil, fmt.Errorf("%w: platform_fee_pct must not be negative", ErrInvalidTenant)
	}
	if input.ExchangeFeePct < 0 {
		return nil, fmt.Errorf("%w: exchange_fee_pct must not be negative", ErrInvalidTenant)
	}

	now := time.Now().UTC()
	return &Tenant{
		ID:             uuid.New(),
		Name:           input.Name,
		Slug:           input.Slug,
		Domain:         input.Domain,
		LogoURL:        input.LogoURL,
		PrimaryColor:   input.PrimaryColor,
		PlatformFeePct: input.PlatformFeePct,
		ExchangeFeePct: input.ExchangeFeePct,
		IsActive:       true,
		CreatedAt:      now,
		UpdatedAt:      now,
	}, nil
}
