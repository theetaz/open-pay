package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"
)

var (
	ErrInvalidAPIKey  = errors.New("invalid API key format")
	ErrInvalidEnv     = errors.New("environment must be 'live' or 'test'")
)

// ParseAPIKey splits a compound API key into its key ID and secret parts.
// Format: "ak_{env}_{id}.sk_{env}_{secret}"
func ParseAPIKey(apiKey string) (keyID, secret string, err error) {
	if apiKey == "" {
		return "", "", ErrInvalidAPIKey
	}

	parts := strings.SplitN(apiKey, ".", 2)
	if len(parts) != 2 {
		return "", "", ErrInvalidAPIKey
	}

	keyID = parts[0]
	secret = parts[1]

	if !strings.HasPrefix(keyID, "ak_live_") && !strings.HasPrefix(keyID, "ak_test_") {
		return "", "", ErrInvalidAPIKey
	}
	if !strings.HasPrefix(secret, "sk_live_") && !strings.HasPrefix(secret, "sk_test_") {
		return "", "", ErrInvalidAPIKey
	}

	return keyID, secret, nil
}

// GenerateAPIKey creates a new API key pair for the given environment.
func GenerateAPIKey(env string) (string, error) {
	if env != "live" && env != "test" {
		return "", ErrInvalidEnv
	}

	keyID, err := randomHex(16)
	if err != nil {
		return "", fmt.Errorf("generating key ID: %w", err)
	}

	secret, err := randomHex(32)
	if err != nil {
		return "", fmt.Errorf("generating secret: %w", err)
	}

	return fmt.Sprintf("ak_%s_%s.sk_%s_%s", env, keyID, env, secret), nil
}

// SignRequest creates an HMAC-SHA256 signature for an API request.
// The signing key is derived by hashing the secret with SHA256.
// Message format: timestamp + method + path + body (no separators).
func SignRequest(secret, timestamp, method, path, body string) string {
	// Derive signing key
	h := sha256.Sum256([]byte(secret))
	signingKey := h[:]

	// Build message
	message := timestamp + strings.ToUpper(method) + path + body

	// HMAC-SHA256
	mac := hmac.New(sha256.New, signingKey)
	mac.Write([]byte(message))
	return hex.EncodeToString(mac.Sum(nil))
}

// VerifySignature checks if a request signature is valid.
func VerifySignature(secret, timestamp, method, path, body, signature string) bool {
	expected := SignRequest(secret, timestamp, method, path, body)
	return hmac.Equal([]byte(expected), []byte(signature))
}

// CurrentTimestamp returns the current time as Unix milliseconds string.
func CurrentTimestamp() string {
	return strconv.FormatInt(time.Now().UnixMilli(), 10)
}

// IsTimestampValid checks if a timestamp is within the allowed window.
// windowMs is the allowed deviation in milliseconds.
func IsTimestampValid(timestamp string, windowMs int64) bool {
	ts, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return false
	}

	now := time.Now().UnixMilli()
	diff := now - ts
	if diff < 0 {
		diff = -diff
	}

	return diff <= windowMs
}

func randomHex(n int) (string, error) {
	if n <= 0 || n > 64 {
		return "", fmt.Errorf("randomHex: n must be between 1 and 64, got %d", n)
	}
	//nolint:gosec // G115: n is validated above (1-64), so n*8 fits in uint
	max := new(big.Int).Lsh(big.NewInt(1), uint(n*8))
	val, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%0*x", n*2, val), nil
}
