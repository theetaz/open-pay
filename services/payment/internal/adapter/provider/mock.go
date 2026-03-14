package provider

import (
	"context"
	"fmt"
	"sync"

	"github.com/google/uuid"
	"github.com/openlankapay/openlankapay/services/payment/internal/domain"
)

// MockProvider simulates a payment provider for development and testing.
// Special amounts trigger specific behaviors:
// - 0.01: auto-expires
// - 0.02: auto-fails
// - All others: stays INITIATED until SimulatePayment is called
type MockProvider struct {
	mu       sync.RWMutex
	payments map[string]*mockPayment
}

type mockPayment struct {
	amount string
	status domain.PaymentStatus
	txHash string
}

// NewMockProvider creates a new mock payment provider.
func NewMockProvider() *MockProvider {
	return &MockProvider{
		payments: make(map[string]*mockPayment),
	}
}

func (p *MockProvider) Name() string {
	return "TEST"
}

func (p *MockProvider) CreatePayment(_ context.Context, req domain.ProviderPaymentRequest) (*domain.ProviderPaymentResponse, error) {
	payID := fmt.Sprintf("mock_%s", uuid.New().String()[:12])

	var status domain.PaymentStatus
	switch req.Amount {
	case "0.01":
		status = domain.StatusExpired
	case "0.02":
		status = domain.StatusFailed
	default:
		status = domain.StatusInitiated
	}

	p.mu.Lock()
	p.payments[payID] = &mockPayment{
		amount: req.Amount,
		status: status,
	}
	p.mu.Unlock()

	return &domain.ProviderPaymentResponse{
		ProviderPayID: payID,
		QRContent:     fmt.Sprintf("mock-qr://%s?amount=%s", payID, req.Amount),
		CheckoutLink:  fmt.Sprintf("https://mock-checkout.test/%s", payID),
		DeepLink:      fmt.Sprintf("mock-app://pay/%s", payID),
	}, nil
}

func (p *MockProvider) GetPaymentStatus(_ context.Context, providerPayID string) (*domain.ProviderPaymentStatus, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	mp, ok := p.payments[providerPayID]
	if !ok {
		return nil, fmt.Errorf("mock payment %s not found", providerPayID)
	}

	return &domain.ProviderPaymentStatus{
		Status: mp.status,
		TxHash: mp.txHash,
	}, nil
}

// SimulatePayment marks a mock payment as paid (for testing).
func (p *MockProvider) SimulatePayment(providerPayID string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if mp, ok := p.payments[providerPayID]; ok {
		mp.status = domain.StatusPaid
		mp.txHash = fmt.Sprintf("0xmock_%s", uuid.New().String()[:16])
	}
}
