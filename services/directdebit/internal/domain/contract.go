package domain

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

var (
	ErrInvalidContract         = errors.New("invalid direct debit contract")
	ErrInvalidPayment          = errors.New("invalid direct debit payment")
	ErrInvalidScenario         = errors.New("invalid scenario code")
	ErrContractNotFound        = errors.New("contract not found")
	ErrPaymentNotFound         = errors.New("direct debit payment not found")
	ErrScenarioNotFound        = errors.New("scenario code not found")
	ErrInvalidStatusTransition = errors.New("invalid contract status transition")
	ErrContractNotSigned       = errors.New("contract must be SIGNED to execute payment")
	ErrAmountExceedsLimit      = errors.New("amount exceeds single upper limit")
)

// Contract status values.
type ContractStatus = string

const (
	ContractInitiated  ContractStatus = "INITIATED"
	ContractSigned     ContractStatus = "SIGNED"
	ContractTerminated ContractStatus = "TERMINATED"
	ContractExpired    ContractStatus = "EXPIRED"
)

var validContractTransitions = map[ContractStatus][]ContractStatus{
	ContractInitiated: {ContractSigned, ContractExpired},
	ContractSigned:    {ContractTerminated, ContractExpired},
}

// Payment status values.
type PaymentStatus = string

const (
	PaymentInitiated PaymentStatus = "INITIATED"
	PaymentPaid      PaymentStatus = "PAID"
	PaymentFailed    PaymentStatus = "FAILED"
	PaymentRefunded  PaymentStatus = "REFUNDED"
)

// PaymentProvider values.
const (
	ProviderBinancePay = "BINANCE_PAY"
	ProviderBybitPay   = "BYBIT_PAY"
	ProviderKuCoinPay  = "KUCOIN_PAY"
)

// ScenarioCode defines a provider-specific pre-authorization type with limits.
type ScenarioCode struct {
	ID              uuid.UUID
	ScenarioID      string
	ScenarioName    string
	PaymentProvider string
	MaxLimit        decimal.Decimal
	IsActive        bool
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// Contract represents a direct debit pre-authorization agreement.
type Contract struct {
	ID                   uuid.UUID
	MerchantID           uuid.UUID
	BranchID             *uuid.UUID
	MerchantContractCode string
	ServiceName          string
	ScenarioID           uuid.UUID
	PaymentProvider      string
	Currency             string
	SingleUpperLimit     decimal.Decimal
	Status               ContractStatus
	PreContractID        string
	ContractID           string
	BizID                string
	OpenUserID           string
	MerchantAccountNo    string
	QRContent            string
	DeepLink             string
	WebhookURL           string
	ReturnURL            string
	CancelURL            string
	Periodic             bool
	ContractEndTime      *time.Time
	RequestExpireTime    *time.Time
	TerminationWay       *int
	TerminationTime      *time.Time
	TerminationNotes     string
	PaymentCount         int
	TotalAmountCharged   decimal.Decimal
	LastPaymentAt        *time.Time
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

// NewContract creates a validated direct debit contract.
func NewContract(merchantID uuid.UUID, serviceName string, scenarioID uuid.UUID, provider string, singleUpperLimit decimal.Decimal, returnURL, cancelURL string) (*Contract, error) {
	if serviceName == "" {
		return nil, fmt.Errorf("%w: service name is required", ErrInvalidContract)
	}
	if len(serviceName) > 64 {
		return nil, fmt.Errorf("%w: service name must be 64 characters or less", ErrInvalidContract)
	}
	if singleUpperLimit.LessThan(decimal.NewFromFloat(0.01)) {
		return nil, fmt.Errorf("%w: single upper limit must be at least 0.01", ErrInvalidContract)
	}
	if returnURL == "" {
		return nil, fmt.Errorf("%w: return URL is required", ErrInvalidContract)
	}
	if cancelURL == "" {
		return nil, fmt.Errorf("%w: cancel URL is required", ErrInvalidContract)
	}

	now := time.Now().UTC()
	contractCode := fmt.Sprintf("DD-%s", uuid.New().String()[:8])

	return &Contract{
		ID:                   uuid.New(),
		MerchantID:           merchantID,
		MerchantContractCode: contractCode,
		ServiceName:          serviceName,
		ScenarioID:           scenarioID,
		PaymentProvider:      provider,
		Currency:             "USDT",
		SingleUpperLimit:     singleUpperLimit,
		Status:               ContractInitiated,
		CreatedAt:            now,
		UpdatedAt:            now,
	}, nil
}

// TransitionTo moves the contract to a new status if valid.
func (c *Contract) TransitionTo(to ContractStatus) error {
	allowed, ok := validContractTransitions[c.Status]
	if !ok {
		return fmt.Errorf("%w: no transitions from %s", ErrInvalidStatusTransition, c.Status)
	}
	for _, a := range allowed {
		if a == to {
			c.Status = to
			c.UpdatedAt = time.Now().UTC()
			return nil
		}
	}
	return fmt.Errorf("%w: cannot transition from %s to %s", ErrInvalidStatusTransition, c.Status, to)
}

// Terminate marks the contract as terminated.
func (c *Contract) Terminate(notes string) error {
	if err := c.TransitionTo(ContractTerminated); err != nil {
		return err
	}
	now := time.Now().UTC()
	way := 1
	c.TerminationWay = &way
	c.TerminationTime = &now
	c.TerminationNotes = notes
	return nil
}

// CanExecutePayment checks if the contract is in a state that allows payments.
func (c *Contract) CanExecutePayment(amount decimal.Decimal) error {
	if c.Status != ContractSigned {
		return ErrContractNotSigned
	}
	if amount.GreaterThan(c.SingleUpperLimit) {
		return fmt.Errorf("%w: %s exceeds limit %s", ErrAmountExceedsLimit, amount.String(), c.SingleUpperLimit.String())
	}
	return nil
}

// Payment represents a charge executed against a signed direct debit contract.
type Payment struct {
	ID              uuid.UUID
	ContractID      uuid.UUID
	MerchantID      uuid.UUID
	PayID           string
	PaymentNo       string
	Amount          decimal.Decimal
	Currency        string
	Status          PaymentStatus
	ProductName     string
	ProductDetail   string
	PaymentProvider string
	WebhookURL      string
	// Fee breakdown
	GrossAmountUSDT    decimal.Decimal
	ExchangeFeePct     decimal.Decimal
	ExchangeFeeUSDT    decimal.Decimal
	PlatformFeePct     decimal.Decimal
	PlatformFeeUSDT    decimal.Decimal
	TotalFeesUSDT      decimal.Decimal
	NetAmountUSDT      decimal.Decimal
	// Customer billing
	CustomerFirstName string
	CustomerLastName  string
	CustomerEmail     string
	CustomerPhone     string
	CustomerAddress   string
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// NewPayment creates a validated direct debit payment against a contract.
func NewPayment(contract *Contract, amount decimal.Decimal, productName string) (*Payment, error) {
	if err := contract.CanExecutePayment(amount); err != nil {
		return nil, err
	}
	if productName == "" {
		return nil, fmt.Errorf("%w: product name is required", ErrInvalidPayment)
	}

	now := time.Now().UTC()
	paymentNo := fmt.Sprintf("DDP-%s", uuid.New().String()[:12])

	// Calculate fees (0.5% exchange, 1.5% platform)
	exchangeFeePct := decimal.NewFromFloat(0.5)
	platformFeePct := decimal.NewFromFloat(1.5)
	exchangeFee := amount.Mul(exchangeFeePct).Div(decimal.NewFromInt(100))
	platformFee := amount.Mul(platformFeePct).Div(decimal.NewFromInt(100))
	totalFees := exchangeFee.Add(platformFee)
	netAmount := amount.Sub(totalFees)

	return &Payment{
		ID:              uuid.New(),
		ContractID:      contract.ID,
		MerchantID:      contract.MerchantID,
		PaymentNo:       paymentNo,
		Amount:          amount,
		Currency:        contract.Currency,
		Status:          PaymentInitiated,
		ProductName:     productName,
		PaymentProvider: contract.PaymentProvider,
		WebhookURL:      contract.WebhookURL,
		GrossAmountUSDT: amount,
		ExchangeFeePct:  exchangeFeePct,
		ExchangeFeeUSDT: exchangeFee,
		PlatformFeePct:  platformFeePct,
		PlatformFeeUSDT: platformFee,
		TotalFeesUSDT:   totalFees,
		NetAmountUSDT:   netAmount,
		CreatedAt:       now,
		UpdatedAt:       now,
	}, nil
}

// MarkPaid transitions the payment to PAID status.
func (p *Payment) MarkPaid(payID string) {
	p.Status = PaymentPaid
	p.PayID = payID
	p.UpdatedAt = time.Now().UTC()
}

// MarkFailed transitions the payment to FAILED status.
func (p *Payment) MarkFailed() {
	p.Status = PaymentFailed
	p.UpdatedAt = time.Now().UTC()
}
