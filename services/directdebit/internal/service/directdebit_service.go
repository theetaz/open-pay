package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/openlankapay/openlankapay/services/directdebit/internal/domain"
)

// ScenarioRepository defines the data access contract for scenario codes.
type ScenarioRepository interface {
	Create(ctx context.Context, s *domain.ScenarioCode) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.ScenarioCode, error)
	List(ctx context.Context, provider string, activeOnly bool) ([]*domain.ScenarioCode, error)
}

// ContractRepository defines the data access contract for direct debit contracts.
type ContractRepository interface {
	Create(ctx context.Context, c *domain.Contract) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Contract, error)
	ListByMerchant(ctx context.Context, merchantID uuid.UUID, status string, page, limit int) ([]*domain.Contract, int, error)
	Update(ctx context.Context, c *domain.Contract) error
}

// PaymentRepository defines the data access contract for direct debit payments.
type PaymentRepository interface {
	Create(ctx context.Context, p *domain.Payment) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Payment, error)
	ListByContract(ctx context.Context, contractID uuid.UUID) ([]*domain.Payment, error)
	Update(ctx context.Context, p *domain.Payment) error
}

// DirectDebitService orchestrates direct debit operations.
type DirectDebitService struct {
	scenarios ScenarioRepository
	contracts ContractRepository
	payments  PaymentRepository
}

// NewDirectDebitService creates a new direct debit service.
func NewDirectDebitService(scenarios ScenarioRepository, contracts ContractRepository, payments PaymentRepository) *DirectDebitService {
	return &DirectDebitService{
		scenarios: scenarios,
		contracts: contracts,
		payments:  payments,
	}
}

// ListScenarioCodes returns available scenario codes.
func (s *DirectDebitService) ListScenarioCodes(ctx context.Context, provider string, activeOnly bool) ([]*domain.ScenarioCode, error) {
	return s.scenarios.List(ctx, provider, activeOnly)
}

// CreateContract creates a new direct debit contract.
func (s *DirectDebitService) CreateContract(ctx context.Context, merchantID uuid.UUID, input CreateContractInput) (*domain.Contract, error) {
	// Validate scenario exists
	scenario, err := s.scenarios.GetByID(ctx, input.ScenarioID)
	if err != nil {
		return nil, err
	}

	// Validate single upper limit against scenario max limit
	if input.SingleUpperLimit.GreaterThan(scenario.MaxLimit) {
		return nil, fmt.Errorf("%w: single upper limit %s exceeds scenario max %s", domain.ErrInvalidContract, input.SingleUpperLimit.String(), scenario.MaxLimit.String())
	}

	contract, err := domain.NewContract(merchantID, input.ServiceName, input.ScenarioID, scenario.PaymentProvider, input.SingleUpperLimit, input.ReturnURL, input.CancelURL)
	if err != nil {
		return nil, err
	}

	if input.MerchantContractCode != "" {
		contract.MerchantContractCode = input.MerchantContractCode
	}
	if input.BranchID != nil {
		contract.BranchID = input.BranchID
	}
	contract.WebhookURL = input.WebhookURL

	// In a real implementation, this would call the payment provider API
	// to create the pre-authorization and get QR content + deep link.
	// For now, we generate mock values.
	contract.QRContent = fmt.Sprintf("dd-contract://%s", contract.ID.String())
	contract.DeepLink = fmt.Sprintf("https://pay.example.com/dd/%s", contract.ID.String())
	expire := time.Now().UTC().Add(30 * time.Minute)
	contract.RequestExpireTime = &expire

	if err := s.contracts.Create(ctx, contract); err != nil {
		return nil, fmt.Errorf("storing contract: %w", err)
	}

	return contract, nil
}

// GetContract returns a contract by ID.
func (s *DirectDebitService) GetContract(ctx context.Context, id uuid.UUID) (*domain.Contract, error) {
	return s.contracts.GetByID(ctx, id)
}

// ListContracts returns contracts for a merchant with pagination.
func (s *DirectDebitService) ListContracts(ctx context.Context, merchantID uuid.UUID, status string, page, limit int) ([]*domain.Contract, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	return s.contracts.ListByMerchant(ctx, merchantID, status, page, limit)
}

// SyncContractStatus queries the payment provider for the latest contract status.
func (s *DirectDebitService) SyncContractStatus(ctx context.Context, id uuid.UUID) (*domain.Contract, error) {
	contract, err := s.contracts.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// In production, call the payment provider API to check contract status.
	// For sandbox: auto-sign contracts that are INITIATED.
	if contract.Status == domain.ContractInitiated {
		if err := contract.TransitionTo(domain.ContractSigned); err != nil {
			return nil, err
		}
		contract.ContractID = fmt.Sprintf("prov-contract-%s", uuid.New().String()[:8])
		contract.OpenUserID = fmt.Sprintf("user-%s", uuid.New().String()[:8])
		if err := s.contracts.Update(ctx, contract); err != nil {
			return nil, fmt.Errorf("updating contract: %w", err)
		}
	}

	return contract, nil
}

// TerminateContract terminates a signed contract.
func (s *DirectDebitService) TerminateContract(ctx context.Context, id uuid.UUID, notes string) (*domain.Contract, error) {
	contract, err := s.contracts.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := contract.Terminate(notes); err != nil {
		return nil, err
	}

	if err := s.contracts.Update(ctx, contract); err != nil {
		return nil, fmt.Errorf("updating contract: %w", err)
	}

	return contract, nil
}

// ExecutePayment charges an amount against a signed contract.
func (s *DirectDebitService) ExecutePayment(ctx context.Context, contractID uuid.UUID, input ExecutePaymentInput) (*domain.Payment, error) {
	contract, err := s.contracts.GetByID(ctx, contractID)
	if err != nil {
		return nil, err
	}

	payment, err := domain.NewPayment(contract, input.Amount, input.ProductName)
	if err != nil {
		return nil, err
	}
	payment.ProductDetail = input.ProductDetail
	if input.WebhookURL != "" {
		payment.WebhookURL = input.WebhookURL
	}
	payment.CustomerFirstName = input.CustomerFirstName
	payment.CustomerLastName = input.CustomerLastName
	payment.CustomerEmail = input.CustomerEmail
	payment.CustomerPhone = input.CustomerPhone
	payment.CustomerAddress = input.CustomerAddress

	// In production, call provider API to execute the payment.
	// For sandbox: auto-complete payments.
	payment.MarkPaid(fmt.Sprintf("prov-pay-%s", uuid.New().String()[:8]))

	if err := s.payments.Create(ctx, payment); err != nil {
		return nil, fmt.Errorf("storing payment: %w", err)
	}

	// Update contract stats
	contract.PaymentCount++
	contract.TotalAmountCharged = contract.TotalAmountCharged.Add(payment.Amount)
	now := time.Now().UTC()
	contract.LastPaymentAt = &now
	contract.UpdatedAt = now
	if err := s.contracts.Update(ctx, contract); err != nil {
		return nil, fmt.Errorf("updating contract stats: %w", err)
	}

	return payment, nil
}

// CreateContractInput holds input for creating a direct debit contract.
type CreateContractInput struct {
	MerchantContractCode string
	BranchID             *uuid.UUID
	ServiceName          string
	ScenarioID           uuid.UUID
	SingleUpperLimit     decimal.Decimal
	WebhookURL           string
	ReturnURL            string
	CancelURL            string
}

// ExecutePaymentInput holds input for executing a payment against a contract.
type ExecutePaymentInput struct {
	Amount            decimal.Decimal
	ProductName       string
	ProductDetail     string
	WebhookURL        string
	CustomerFirstName string
	CustomerLastName  string
	CustomerEmail     string
	CustomerPhone     string
	CustomerAddress   string
}
