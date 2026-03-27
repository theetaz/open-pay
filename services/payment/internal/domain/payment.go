package domain

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

var (
	ErrInvalidPayment          = errors.New("invalid payment data")
	ErrInvalidStatusTransition = errors.New("invalid payment status transition")
	ErrPaymentNotFound         = errors.New("payment not found")
	ErrPaymentExpired          = errors.New("payment has expired")
)

// PaymentStatus represents the state of a payment.
type PaymentStatus string

const (
	StatusInitiated  PaymentStatus = "INITIATED"
	StatusUserReview PaymentStatus = "USER_REVIEW"
	StatusPaid       PaymentStatus = "PAID"
	StatusExpired    PaymentStatus = "EXPIRED"
	StatusFailed     PaymentStatus = "FAILED"
)

// Valid currencies for payments.
var validCurrencies = map[string]bool{
	"USDT": true,
	"USDC": true,
	"BTC":  true,
	"ETH":  true,
	"BNB":  true,
	"LKR":  true,
}

// Valid payment providers.
var validProviders = map[string]bool{
	"BYBIT":   true,
	"BINANCE": true,
	"KUCOIN":  true,
	"TEST":    true,
}

// Valid status transitions.
var validTransitions = map[PaymentStatus][]PaymentStatus{
	StatusInitiated:  {StatusUserReview, StatusPaid, StatusExpired, StatusFailed},
	StatusUserReview: {StatusPaid, StatusFailed, StatusExpired},
}

// Default payment expiration.
const DefaultExpiration = 15 * time.Minute

// GoodItem represents a product or service in a payment.
type GoodItem struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	MccCode     string `json:"mccCode,omitempty"`
}

// CustomerBilling holds customer billing details.
type CustomerBilling struct {
	FirstName  string `json:"firstName"`
	LastName   string `json:"lastName"`
	Email      string `json:"email"`
	Phone      string `json:"phone"`
	Address    string `json:"address"`
	City       string `json:"city,omitempty"`
	PostalCode string `json:"postalCode,omitempty"`
	Country    string `json:"country,omitempty"`
}

// CreatePaymentInput holds the data needed to create a payment.
type CreatePaymentInput struct {
	MerchantID        uuid.UUID
	BranchID          *uuid.UUID
	Amount            decimal.Decimal
	Currency          string
	Provider          string
	MerchantTradeNo   string
	WebhookURL        string
	SuccessURL        string
	CancelURL         string
	ExpireTime        *time.Time
	CustomerEmail     string
	CustomerFirstName string
	CustomerLastName  string
	CustomerPhone     string
	CustomerAddress   string
	Goods             []GoodItem
}

// Payment represents a payment order.
type Payment struct {
	ID              uuid.UUID
	MerchantID      uuid.UUID
	BranchID        *uuid.UUID
	PaymentNo       string
	MerchantTradeNo string
	Amount          decimal.Decimal
	Currency        string
	AmountUSDT      decimal.Decimal
	ExchangeRateSnapshot *decimal.Decimal
	ExchangeFeePct  decimal.Decimal
	ExchangeFeeUSDT decimal.Decimal
	PlatformFeePct  decimal.Decimal
	PlatformFeeUSDT decimal.Decimal
	TotalFeesUSDT   decimal.Decimal
	NetAmountUSDT   decimal.Decimal
	Provider        string
	ProviderPayID   string
	QRContent       string
	CheckoutLink    string
	DeepLink        string
	Status          PaymentStatus
	CustomerEmail     string
	CustomerFirstName string
	CustomerLastName  string
	CustomerPhone     string
	CustomerAddress   string
	Goods             []GoodItem
	WebhookURL      string
	SuccessURL      string
	CancelURL       string
	TxHash          string
	BlockNumber     int64
	WalletAddress   string
	// LKR-specific fee fields (populated when currency is LKR)
	LKRAmount       *decimal.Decimal
	LKRExchangeFee  *decimal.Decimal
	LKRPlatformFee  *decimal.Decimal
	LKRTotalFees    *decimal.Decimal
	LKRNetAmount    *decimal.Decimal
	ExpireTime      time.Time
	PaidAt          *time.Time
	FailedAt        *time.Time
	IdempotencyKey  string
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       *time.Time
}

// NewPayment creates a validated Payment from input.
func NewPayment(input CreatePaymentInput) (*Payment, error) {
	if input.MerchantID == uuid.Nil {
		return nil, fmt.Errorf("%w: merchant ID is required", ErrInvalidPayment)
	}
	if input.Amount.LessThanOrEqual(decimal.Zero) {
		return nil, fmt.Errorf("%w: amount must be positive", ErrInvalidPayment)
	}
	if !validCurrencies[input.Currency] {
		return nil, fmt.Errorf("%w: unsupported currency %s", ErrInvalidPayment, input.Currency)
	}
	if !validProviders[input.Provider] {
		return nil, fmt.Errorf("%w: unsupported provider %s", ErrInvalidPayment, input.Provider)
	}

	now := time.Now().UTC()
	expireTime := now.Add(DefaultExpiration)
	if input.ExpireTime != nil {
		expireTime = *input.ExpireTime
	}

	return &Payment{
		ID:                uuid.New(),
		MerchantID:        input.MerchantID,
		BranchID:          input.BranchID,
		PaymentNo:         generatePaymentNo(now),
		MerchantTradeNo:   input.MerchantTradeNo,
		Amount:            input.Amount,
		Currency:          input.Currency,
		AmountUSDT:        input.Amount, // Will be overwritten if LKR
		Provider:          input.Provider,
		Status:            StatusInitiated,
		CustomerEmail:     input.CustomerEmail,
		CustomerFirstName: input.CustomerFirstName,
		CustomerLastName:  input.CustomerLastName,
		CustomerPhone:     input.CustomerPhone,
		CustomerAddress:   input.CustomerAddress,
		Goods:             input.Goods,
		WebhookURL:        input.WebhookURL,
		SuccessURL:        input.SuccessURL,
		CancelURL:         input.CancelURL,
		ExpireTime:        expireTime,
		CreatedAt:         now,
		UpdatedAt:         now,
	}, nil
}

// TransitionTo moves the payment to a new status if the transition is valid.
func (p *Payment) TransitionTo(to PaymentStatus) error {
	allowed, ok := validTransitions[p.Status]
	if !ok {
		return fmt.Errorf("%w: no transitions from %s", ErrInvalidStatusTransition, p.Status)
	}

	for _, s := range allowed {
		if s == to {
			p.Status = to
			p.UpdatedAt = time.Now().UTC()
			return nil
		}
	}

	return fmt.Errorf("%w: cannot transition from %s to %s", ErrInvalidStatusTransition, p.Status, to)
}

// MarkPaid transitions the payment to PAID and records transaction details.
func (p *Payment) MarkPaid(txHash string) error {
	if err := p.TransitionTo(StatusPaid); err != nil {
		return err
	}
	now := time.Now().UTC()
	p.PaidAt = &now
	p.TxHash = txHash
	return nil
}

// IsExpired checks if the payment has passed its expiration time
// and is still in a non-terminal state.
func (p *Payment) IsExpired() bool {
	if p.Status == StatusPaid || p.Status == StatusFailed || p.Status == StatusExpired {
		return false
	}
	return time.Now().After(p.ExpireTime)
}

// SetFees calculates and sets the fee breakdown.
// Fee rates are percentages (e.g., 0.5 means 0.5%).
func (p *Payment) SetFees(exchangeFeeRate, platformFeeRate decimal.Decimal) {
	hundred := decimal.NewFromInt(100)

	p.ExchangeFeePct = exchangeFeeRate
	p.ExchangeFeeUSDT = p.AmountUSDT.Mul(exchangeFeeRate).Div(hundred)
	p.PlatformFeePct = platformFeeRate
	p.PlatformFeeUSDT = p.AmountUSDT.Mul(platformFeeRate).Div(hundred)
	p.TotalFeesUSDT = p.ExchangeFeeUSDT.Add(p.PlatformFeeUSDT)
	p.NetAmountUSDT = p.AmountUSDT.Sub(p.TotalFeesUSDT)

	// Calculate LKR equivalents when currency is LKR
	if p.Currency == "LKR" && p.ExchangeRateSnapshot != nil {
		lkrAmount := p.Amount
		lkrExchFee := lkrAmount.Mul(exchangeFeeRate).Div(hundred)
		lkrPlatFee := lkrAmount.Mul(platformFeeRate).Div(hundred)
		lkrTotalFees := lkrExchFee.Add(lkrPlatFee)
		lkrNet := lkrAmount.Sub(lkrTotalFees)
		p.LKRAmount = &lkrAmount
		p.LKRExchangeFee = &lkrExchFee
		p.LKRPlatformFee = &lkrPlatFee
		p.LKRTotalFees = &lkrTotalFees
		p.LKRNetAmount = &lkrNet
	}
}

// SetExchangeRate sets the exchange rate and converts LKR to USDT.
func (p *Payment) SetExchangeRate(rate decimal.Decimal) {
	p.ExchangeRateSnapshot = &rate
	if p.Currency == "LKR" {
		p.AmountUSDT = p.Amount.Div(rate)
	}
}

func generatePaymentNo(t time.Time) string {
	short := uuid.New().String()[:8]
	return fmt.Sprintf("PAY-%s-%s", t.Format("20060102"), short)
}
