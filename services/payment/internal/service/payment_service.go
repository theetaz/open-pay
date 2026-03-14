package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/openlankapay/openlankapay/services/payment/internal/domain"
)

// PaymentRepository defines the data access contract.
type PaymentRepository interface {
	Create(ctx context.Context, payment *domain.Payment) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Payment, error)
	Update(ctx context.Context, payment *domain.Payment) error
	List(ctx context.Context, merchantID uuid.UUID, params ListParams) ([]*domain.Payment, int, error)
}

// EventPublisher defines the contract for publishing domain events.
type EventPublisher interface {
	Publish(ctx context.Context, subject string, data any) error
}

// ListParams holds pagination parameters.
type ListParams struct {
	Page    int
	PerPage int
	Status  *domain.PaymentStatus
}

// CreatePaymentInput holds the data needed to create a payment.
type CreatePaymentInput struct {
	MerchantID      uuid.UUID
	BranchID        *uuid.UUID
	Amount          decimal.Decimal
	Currency        string
	Provider        string
	MerchantTradeNo string
	WebhookURL      string
	CustomerEmail   string
}

// PaymentService orchestrates payment operations.
type PaymentService struct {
	repo      PaymentRepository
	providers map[string]domain.PaymentProvider
	events    EventPublisher
}

// NewPaymentService creates a new PaymentService.
func NewPaymentService(repo PaymentRepository, providers map[string]domain.PaymentProvider, events EventPublisher) *PaymentService {
	return &PaymentService{
		repo:      repo,
		providers: providers,
		events:    events,
	}
}

// CreatePayment creates a new payment and initiates it with the provider.
func (s *PaymentService) CreatePayment(ctx context.Context, input CreatePaymentInput) (*domain.Payment, error) {
	prov, ok := s.providers[input.Provider]
	if !ok {
		return nil, fmt.Errorf("unsupported provider: %s", input.Provider)
	}

	payment, err := domain.NewPayment(domain.CreatePaymentInput{
		MerchantID:      input.MerchantID,
		BranchID:        input.BranchID,
		Amount:          input.Amount,
		Currency:        input.Currency,
		Provider:        input.Provider,
		MerchantTradeNo: input.MerchantTradeNo,
		WebhookURL:      input.WebhookURL,
		CustomerEmail:   input.CustomerEmail,
	})
	if err != nil {
		return nil, err
	}

	// Set fees (default rates)
	payment.SetFees(decimal.NewFromFloat(0.5), decimal.NewFromFloat(1.5))

	// Create payment with provider
	provResp, err := prov.CreatePayment(ctx, domain.ProviderPaymentRequest{
		Amount:   payment.AmountUSDT.String(),
		Currency: "USDT",
		OrderID:  payment.ID.String(),
	})
	if err != nil {
		return nil, fmt.Errorf("provider error: %w", err)
	}

	payment.ProviderPayID = provResp.ProviderPayID
	payment.QRContent = provResp.QRContent
	payment.CheckoutLink = provResp.CheckoutLink
	payment.DeepLink = provResp.DeepLink

	if err := s.repo.Create(ctx, payment); err != nil {
		return nil, fmt.Errorf("storing payment: %w", err)
	}

	_ = s.events.Publish(ctx, "payment.initiated", payment)

	return payment, nil
}

// GetPayment retrieves a payment by ID.
func (s *PaymentService) GetPayment(ctx context.Context, id uuid.UUID) (*domain.Payment, error) {
	return s.repo.GetByID(ctx, id)
}

// ListPayments returns paginated payments for a merchant.
func (s *PaymentService) ListPayments(ctx context.Context, merchantID uuid.UUID, params ListParams) ([]*domain.Payment, int, error) {
	return s.repo.List(ctx, merchantID, params)
}

// HandleProviderCallback processes a payment status update from a provider.
func (s *PaymentService) HandleProviderCallback(ctx context.Context, paymentID uuid.UUID) error {
	payment, err := s.repo.GetByID(ctx, paymentID)
	if err != nil {
		return err
	}

	prov, ok := s.providers[payment.Provider]
	if !ok {
		return fmt.Errorf("unknown provider: %s", payment.Provider)
	}

	status, err := prov.GetPaymentStatus(ctx, payment.ProviderPayID)
	if err != nil {
		return fmt.Errorf("checking provider status: %w", err)
	}

	switch status.Status {
	case domain.StatusPaid:
		if err := payment.MarkPaid(status.TxHash); err != nil {
			return err
		}
		_ = s.events.Publish(ctx, "payment.paid", payment)
	case domain.StatusExpired:
		if err := payment.TransitionTo(domain.StatusExpired); err != nil {
			return err
		}
		_ = s.events.Publish(ctx, "payment.expired", payment)
	case domain.StatusFailed:
		if err := payment.TransitionTo(domain.StatusFailed); err != nil {
			return err
		}
		_ = s.events.Publish(ctx, "payment.failed", payment)
	default:
		return nil // No status change
	}

	return s.repo.Update(ctx, payment)
}
