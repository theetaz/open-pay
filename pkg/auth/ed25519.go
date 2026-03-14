package auth

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"errors"
)

var ErrInvalidSignature = errors.New("invalid signature")

// GenerateED25519KeyPair creates a new ED25519 key pair.
func GenerateED25519KeyPair() (ed25519.PublicKey, ed25519.PrivateKey, error) {
	return ed25519.GenerateKey(rand.Reader)
}

// SignWebhook creates an ED25519 signature for a webhook payload.
// Message format: timestamp + payload (no separator).
// Returns base64-encoded signature.
func SignWebhook(privateKey ed25519.PrivateKey, timestamp string, payload []byte) (string, error) {
	message := append([]byte(timestamp), payload...)
	signature := ed25519.Sign(privateKey, message)
	return base64.StdEncoding.EncodeToString(signature), nil
}

// VerifyWebhook checks an ED25519 signature against a webhook payload.
// signature should be base64-encoded.
func VerifyWebhook(publicKey ed25519.PublicKey, timestamp string, payload []byte, signature string) (bool, error) {
	sigBytes, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		return false, err
	}

	message := append([]byte(timestamp), payload...)
	return ed25519.Verify(publicKey, message, sigBytes), nil
}

// EncodePublicKey encodes an ED25519 public key to base64.
func EncodePublicKey(pub ed25519.PublicKey) string {
	return base64.StdEncoding.EncodeToString(pub)
}

// DecodePublicKey decodes a base64-encoded ED25519 public key.
func DecodePublicKey(encoded string) (ed25519.PublicKey, error) {
	b, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, err
	}
	if len(b) != ed25519.PublicKeySize {
		return nil, errors.New("invalid public key size")
	}
	return ed25519.PublicKey(b), nil
}

// EncodePrivateKey encodes an ED25519 private key to base64.
func EncodePrivateKey(priv ed25519.PrivateKey) string {
	return base64.StdEncoding.EncodeToString(priv)
}

// DecodePrivateKey decodes a base64-encoded ED25519 private key.
func DecodePrivateKey(encoded string) (ed25519.PrivateKey, error) {
	b, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, err
	}
	if len(b) != ed25519.PrivateKeySize {
		return nil, errors.New("invalid private key size")
	}
	return ed25519.PrivateKey(b), nil
}
