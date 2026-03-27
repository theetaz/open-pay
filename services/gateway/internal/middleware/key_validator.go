package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// RemoteKeyValidator validates API keys by calling the merchant service's internal endpoint.
// It caches HMAC keys to avoid calling the merchant service on every request.
type RemoteKeyValidator struct {
	merchantServiceURL string
	client             *http.Client
	cache              sync.Map // keyID → cachedKey
	cacheTTL           time.Duration
}

type cachedKey struct {
	hmacKey    string
	merchantID string
	allowedIPs []string
	expiresAt  time.Time
}

// NewRemoteKeyValidator creates a new validator that calls the merchant service.
func NewRemoteKeyValidator(merchantServiceURL string) *RemoteKeyValidator {
	return &RemoteKeyValidator{
		merchantServiceURL: merchantServiceURL,
		client:             &http.Client{Timeout: 5 * time.Second},
		cacheTTL:           2 * time.Minute,
	}
}

// GetSecretByKeyID returns the HMAC signing key for the given API key ID.
// This is called by the HMACAuth middleware to verify request signatures.
func (v *RemoteKeyValidator) GetSecretByKeyID(keyID string) (string, error) {
	// Check cache first
	if cached, ok := v.cache.Load(keyID); ok {
		ck := cached.(cachedKey)
		if time.Now().Before(ck.expiresAt) {
			return ck.hmacKey, nil
		}
		v.cache.Delete(keyID)
	}

	// Call merchant service
	body, _ := json.Marshal(map[string]string{"keyId": keyID})
	resp, err := v.client.Post(
		v.merchantServiceURL+"/internal/api-keys/validate",
		"application/json",
		bytes.NewReader(body),
	)
	if err != nil {
		return "", fmt.Errorf("calling merchant service: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return "", ErrKeyNotFound
	}

	var result struct {
		Data struct {
			HMACKey    string   `json:"hmacKey"`
			MerchantID string   `json:"merchantId"`
			AllowedIPs []string `json:"allowedIps"`
		} `json:"data"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("parsing response: %w", err)
	}

	if result.Data.HMACKey == "" {
		return "", ErrKeyNotFound
	}

	// Cache the result
	v.cache.Store(keyID, cachedKey{
		hmacKey:    result.Data.HMACKey,
		merchantID: result.Data.MerchantID,
		allowedIPs: result.Data.AllowedIPs,
		expiresAt:  time.Now().Add(v.cacheTTL),
	})

	return result.Data.HMACKey, nil
}

// GetMerchantID returns the cached merchant ID for a previously validated key.
func (v *RemoteKeyValidator) GetMerchantID(keyID string) string {
	if cached, ok := v.cache.Load(keyID); ok {
		return cached.(cachedKey).merchantID
	}
	return ""
}

// GetAllowedIPs returns the cached allowed IPs for a previously validated key.
// Returns nil if no whitelist is configured (all IPs allowed).
func (v *RemoteKeyValidator) GetAllowedIPs(keyID string) []string {
	if cached, ok := v.cache.Load(keyID); ok {
		return cached.(cachedKey).allowedIPs
	}
	return nil
}
