package domain

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Director represents a company director for KYC verification.
type Director struct {
	ID                uuid.UUID
	MerchantID        uuid.UUID
	Email             string
	FullName          string
	DateOfBirth       *time.Time
	NICPassportNumber string
	Phone             string
	Address           string
	DocumentObjectKey string
	DocumentFilename  string
	VerificationToken string
	TokenExpiresAt    time.Time
	Status            string
	ConsentedAt       *time.Time
	VerifiedAt        *time.Time
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

const (
	DirectorStatusPending  = "PENDING"
	DirectorStatusVerified = "VERIFIED"
	MaxDirectorsPerMerchant = 10
)

// NewDirector creates a new Director with a generated verification token.
func NewDirector(merchantID uuid.UUID, email string) (*Director, error) {
	if email == "" {
		return nil, fmt.Errorf("%w: director email is required", ErrInvalidDirector)
	}
	if err := validateEmail(email); err != nil {
		return nil, fmt.Errorf("%w: %s", ErrInvalidDirector, err.Error())
	}

	token, err := generateToken()
	if err != nil {
		return nil, fmt.Errorf("generating verification token: %w", err)
	}

	now := time.Now().UTC()
	return &Director{
		ID:                uuid.New(),
		MerchantID:        merchantID,
		Email:             email,
		VerificationToken: token,
		TokenExpiresAt:    now.Add(7 * 24 * time.Hour),
		Status:            DirectorStatusPending,
		CreatedAt:         now,
		UpdatedAt:         now,
	}, nil
}

// IsTokenExpired returns true if the verification token has expired.
func (d *Director) IsTokenExpired() bool {
	return time.Now().UTC().After(d.TokenExpiresAt)
}

// RegenerateToken creates a new token with a fresh 7-day expiry.
func (d *Director) RegenerateToken() error {
	token, err := generateToken()
	if err != nil {
		return err
	}
	d.VerificationToken = token
	d.TokenExpiresAt = time.Now().UTC().Add(7 * 24 * time.Hour)
	d.UpdatedAt = time.Now().UTC()
	return nil
}

// Verify marks the director as verified with their submitted details.
func (d *Director) Verify(fullName, nicPassport, phone, address string, dob *time.Time, docKey, docFilename string) error {
	if d.Status == DirectorStatusVerified {
		return fmt.Errorf("%w: director already verified", ErrInvalidDirector)
	}
	if d.IsTokenExpired() {
		return ErrTokenExpired
	}

	now := time.Now().UTC()
	d.FullName = fullName
	d.DateOfBirth = dob
	d.NICPassportNumber = nicPassport
	d.Phone = phone
	d.Address = address
	d.DocumentObjectKey = docKey
	d.DocumentFilename = docFilename
	d.Status = DirectorStatusVerified
	d.ConsentedAt = &now
	d.VerifiedAt = &now
	d.UpdatedAt = now
	return nil
}

func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
