package domain

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// APIKey represents a merchant's API key for authenticating requests.
type APIKey struct {
	ID            uuid.UUID
	MerchantID    uuid.UUID
	KeyID         string // ak_live_xxx or ak_test_xxx (public)
	SecretHash    string // bcrypt hash of the secret
	Name          string
	Environment   string // "live" or "test"
	IsActive      bool
	RevokedAt     *time.Time
	RevokedReason string
	LastUsedAt    *time.Time
	CreatedAt     time.Time
}

// NewAPIKey generates a new API key pair and returns the key entity plus the plain secret.
// The plain secret is only available at creation time; it is stored as a bcrypt hash.
func NewAPIKey(merchantID uuid.UUID, env, name string) (*APIKey, string, error) {
	if env != "live" && env != "test" {
		return nil, "", fmt.Errorf("environment must be 'live' or 'test'")
	}

	keyRand, err := randomHex(16)
	if err != nil {
		return nil, "", fmt.Errorf("generating key ID: %w", err)
	}
	secretRand, err := randomHex(32)
	if err != nil {
		return nil, "", fmt.Errorf("generating secret: %w", err)
	}

	keyID := fmt.Sprintf("ak_%s_%s", env, keyRand)
	plainSecret := fmt.Sprintf("sk_%s_%s", env, secretRand)

	hash, err := bcrypt.GenerateFromPassword([]byte(plainSecret), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", fmt.Errorf("hashing secret: %w", err)
	}

	return &APIKey{
		ID:          uuid.New(),
		MerchantID:  merchantID,
		KeyID:       keyID,
		SecretHash:  string(hash),
		Name:        name,
		Environment: env,
		IsActive:    true,
		CreatedAt:   time.Now().UTC(),
	}, plainSecret, nil
}

// VerifySecret checks if a plain secret matches the stored hash.
func (k *APIKey) VerifySecret(plainSecret string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(k.SecretHash), []byte(plainSecret))
	return err == nil
}

// Revoke marks the API key as inactive.
func (k *APIKey) Revoke(reason string) error {
	if !k.IsActive {
		return ErrKeyAlreadyRevoked
	}
	now := time.Now().UTC()
	k.IsActive = false
	k.RevokedAt = &now
	k.RevokedReason = reason
	return nil
}

func randomHex(n int) (string, error) {
	max := new(big.Int).Lsh(big.NewInt(1), uint(n*8))
	val, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%0*x", n*2, val), nil
}
