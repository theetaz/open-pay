package domain

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

var (
	ErrInsufficientBalance      = errors.New("insufficient balance")
	ErrInvalidWithdrawal        = errors.New("invalid withdrawal")
	ErrInvalidWithdrawalTransition = errors.New("invalid withdrawal status transition")
	ErrBalanceNotFound          = errors.New("balance not found")
	ErrWithdrawalNotFound       = errors.New("withdrawal not found")
	ErrInvalidRefund            = errors.New("invalid refund")
	ErrRefundNotFound           = errors.New("refund not found")
	ErrInvalidRefundTransition  = errors.New("invalid refund status transition")
)

// WithdrawalStatus represents the state of a withdrawal request.
type WithdrawalStatus string

const (
	WithdrawalRequested  WithdrawalStatus = "REQUESTED"
	WithdrawalApproved   WithdrawalStatus = "APPROVED"
	WithdrawalProcessing WithdrawalStatus = "PROCESSING"
	WithdrawalCompleted  WithdrawalStatus = "COMPLETED"
	WithdrawalRejected   WithdrawalStatus = "REJECTED"
)

// MerchantBalance tracks a merchant's running balance.
type MerchantBalance struct {
	ID                uuid.UUID
	MerchantID        uuid.UUID
	AvailableUSDT     decimal.Decimal
	PendingUSDT       decimal.Decimal
	TotalEarnedUSDT   decimal.Decimal
	TotalWithdrawnUSDT decimal.Decimal
	TotalFeesUSDT     decimal.Decimal
	TotalEarnedLKR    decimal.Decimal
	TotalWithdrawnLKR decimal.Decimal
	UpdatedAt         time.Time
}

// NewMerchantBalance creates a zero-balance for a merchant.
func NewMerchantBalance(merchantID uuid.UUID) *MerchantBalance {
	return &MerchantBalance{
		ID:         uuid.New(),
		MerchantID: merchantID,
		UpdatedAt:  time.Now().UTC(),
	}
}

// Debit subtracts a refund amount from the available balance.
func (b *MerchantBalance) Debit(amount decimal.Decimal) error {
	if b.AvailableUSDT.LessThan(amount) {
		return ErrInsufficientBalance
	}
	b.AvailableUSDT = b.AvailableUSDT.Sub(amount)
	b.UpdatedAt = time.Now().UTC()
	return nil
}

// Credit adds a payment's net amount to the available balance.
func (b *MerchantBalance) Credit(netUSDT, feesUSDT decimal.Decimal) {
	b.AvailableUSDT = b.AvailableUSDT.Add(netUSDT)
	b.TotalEarnedUSDT = b.TotalEarnedUSDT.Add(netUSDT)
	b.TotalFeesUSDT = b.TotalFeesUSDT.Add(feesUSDT)
	b.UpdatedAt = time.Now().UTC()
}

// Hold moves an amount from available to pending (for withdrawal).
func (b *MerchantBalance) Hold(amount decimal.Decimal) error {
	if b.AvailableUSDT.LessThan(amount) {
		return ErrInsufficientBalance
	}
	b.AvailableUSDT = b.AvailableUSDT.Sub(amount)
	b.PendingUSDT = b.PendingUSDT.Add(amount)
	b.UpdatedAt = time.Now().UTC()
	return nil
}

// Release moves an amount from pending back to available (rejected withdrawal).
func (b *MerchantBalance) Release(amount decimal.Decimal) {
	b.PendingUSDT = b.PendingUSDT.Sub(amount)
	b.AvailableUSDT = b.AvailableUSDT.Add(amount)
	b.UpdatedAt = time.Now().UTC()
}

// Settle finalizes a withdrawal from the pending balance.
func (b *MerchantBalance) Settle(amount decimal.Decimal) {
	b.PendingUSDT = b.PendingUSDT.Sub(amount)
	b.TotalWithdrawnUSDT = b.TotalWithdrawnUSDT.Add(amount)
	b.UpdatedAt = time.Now().UTC()
}

// Withdrawal represents a merchant's withdrawal request.
type Withdrawal struct {
	ID              uuid.UUID
	MerchantID      uuid.UUID
	AmountUSDT      decimal.Decimal
	ExchangeRate    decimal.Decimal
	AmountLKR       decimal.Decimal
	BankName        string
	BankAccountNo   string
	BankAccountName string
	Status          WithdrawalStatus
	ApprovedBy      *uuid.UUID
	ApprovedAt      *time.Time
	RejectedReason  string
	BankReference   string
	CompletedAt     *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// NewWithdrawal creates a validated withdrawal request.
func NewWithdrawal(merchantID uuid.UUID, amountUSDT, exchangeRate decimal.Decimal, bankName, bankAccountNo, bankAccountName string) (*Withdrawal, error) {
	if amountUSDT.LessThanOrEqual(decimal.Zero) {
		return nil, fmt.Errorf("%w: amount must be positive", ErrInvalidWithdrawal)
	}
	if bankName == "" || bankAccountNo == "" || bankAccountName == "" {
		return nil, fmt.Errorf("%w: bank details are required", ErrInvalidWithdrawal)
	}

	now := time.Now().UTC()
	return &Withdrawal{
		ID:              uuid.New(),
		MerchantID:      merchantID,
		AmountUSDT:      amountUSDT,
		ExchangeRate:    exchangeRate,
		AmountLKR:       amountUSDT.Mul(exchangeRate),
		BankName:        bankName,
		BankAccountNo:   bankAccountNo,
		BankAccountName: bankAccountName,
		Status:          WithdrawalRequested,
		CreatedAt:       now,
		UpdatedAt:       now,
	}, nil
}

// Approve transitions the withdrawal to APPROVED.
func (w *Withdrawal) Approve(adminID uuid.UUID) error {
	if w.Status != WithdrawalRequested {
		return fmt.Errorf("%w: can only approve REQUESTED withdrawals", ErrInvalidWithdrawalTransition)
	}
	now := time.Now().UTC()
	w.Status = WithdrawalApproved
	w.ApprovedBy = &adminID
	w.ApprovedAt = &now
	w.UpdatedAt = now
	return nil
}

// Reject transitions the withdrawal to REJECTED.
func (w *Withdrawal) Reject(reason string) error {
	if w.Status != WithdrawalRequested {
		return fmt.Errorf("%w: can only reject REQUESTED withdrawals", ErrInvalidWithdrawalTransition)
	}
	w.Status = WithdrawalRejected
	w.RejectedReason = reason
	w.UpdatedAt = time.Now().UTC()
	return nil
}

// Complete transitions the withdrawal to COMPLETED with a bank reference.
func (w *Withdrawal) Complete(bankReference string) error {
	if w.Status != WithdrawalApproved {
		return fmt.Errorf("%w: can only complete APPROVED withdrawals", ErrInvalidWithdrawalTransition)
	}
	now := time.Now().UTC()
	w.Status = WithdrawalCompleted
	w.BankReference = bankReference
	w.CompletedAt = &now
	w.UpdatedAt = now
	return nil
}

// RefundStatus represents the state of a refund.
type RefundStatus string

const (
	RefundPending   RefundStatus = "PENDING"
	RefundApproved  RefundStatus = "APPROVED"
	RefundCompleted RefundStatus = "COMPLETED"
	RefundRejected  RefundStatus = "REJECTED"
)

// Refund represents a merchant-initiated refund for a payment.
type Refund struct {
	ID             uuid.UUID
	MerchantID     uuid.UUID
	PaymentID      uuid.UUID
	PaymentNo      string
	AmountUSDT     decimal.Decimal
	Reason         string
	Status         RefundStatus
	ApprovedBy     *uuid.UUID
	ApprovedAt     *time.Time
	RejectedReason string
	CompletedAt    *time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// NewRefund creates a validated refund request.
func NewRefund(merchantID, paymentID uuid.UUID, paymentNo string, amountUSDT decimal.Decimal, reason string) (*Refund, error) {
	if amountUSDT.LessThanOrEqual(decimal.Zero) {
		return nil, fmt.Errorf("%w: amount must be positive", ErrInvalidRefund)
	}
	if reason == "" {
		return nil, fmt.Errorf("%w: reason is required", ErrInvalidRefund)
	}
	if paymentNo == "" {
		return nil, fmt.Errorf("%w: payment number is required", ErrInvalidRefund)
	}

	now := time.Now().UTC()
	return &Refund{
		ID:         uuid.New(),
		MerchantID: merchantID,
		PaymentID:  paymentID,
		PaymentNo:  paymentNo,
		AmountUSDT: amountUSDT,
		Reason:     reason,
		Status:     RefundPending,
		CreatedAt:  now,
		UpdatedAt:  now,
	}, nil
}

// Approve transitions the refund to APPROVED.
func (r *Refund) Approve(adminID uuid.UUID) error {
	if r.Status != RefundPending {
		return fmt.Errorf("%w: can only approve PENDING refunds", ErrInvalidRefundTransition)
	}
	now := time.Now().UTC()
	r.Status = RefundApproved
	r.ApprovedBy = &adminID
	r.ApprovedAt = &now
	r.UpdatedAt = now
	return nil
}

// Complete transitions the refund to COMPLETED.
func (r *Refund) Complete() error {
	if r.Status != RefundApproved {
		return fmt.Errorf("%w: can only complete APPROVED refunds", ErrInvalidRefundTransition)
	}
	now := time.Now().UTC()
	r.Status = RefundCompleted
	r.CompletedAt = &now
	r.UpdatedAt = now
	return nil
}

// Reject transitions the refund to REJECTED.
func (r *Refund) Reject(reason string) error {
	if r.Status != RefundPending {
		return fmt.Errorf("%w: can only reject PENDING refunds", ErrInvalidRefundTransition)
	}
	r.Status = RefundRejected
	r.RejectedReason = reason
	r.UpdatedAt = time.Now().UTC()
	return nil
}
