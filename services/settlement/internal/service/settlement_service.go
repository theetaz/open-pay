package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/openlankapay/openlankapay/services/settlement/internal/domain"
)

// BalanceRepository defines the data access contract for balances.
type BalanceRepository interface {
	GetByMerchantID(ctx context.Context, merchantID uuid.UUID) (*domain.MerchantBalance, error)
	Create(ctx context.Context, b *domain.MerchantBalance) error
	Update(ctx context.Context, b *domain.MerchantBalance) error
}

// WithdrawalRepository defines the data access contract for withdrawals.
type WithdrawalRepository interface {
	Create(ctx context.Context, w *domain.Withdrawal) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Withdrawal, error)
	Update(ctx context.Context, w *domain.Withdrawal) error
	ListByMerchant(ctx context.Context, merchantID uuid.UUID, status *domain.WithdrawalStatus) ([]*domain.Withdrawal, error)
}

// RefundRepository defines the data access contract for refunds.
type RefundRepository interface {
	Create(ctx context.Context, r *domain.Refund) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Refund, error)
	ListByMerchant(ctx context.Context, merchantID uuid.UUID) ([]*domain.Refund, error)
	Update(ctx context.Context, r *domain.Refund) error
}

// SettlementService orchestrates settlement and withdrawal operations.
type SettlementService struct {
	balances    BalanceRepository
	withdrawals WithdrawalRepository
	refunds     RefundRepository
}

// NewSettlementService creates a new SettlementService.
func NewSettlementService(balances BalanceRepository, withdrawals WithdrawalRepository) *SettlementService {
	return &SettlementService{balances: balances, withdrawals: withdrawals}
}

// SetRefundRepository sets the refund repository (optional, added for backwards compatibility).
func (s *SettlementService) SetRefundRepository(r RefundRepository) {
	s.refunds = r
}

// GetBalance returns a merchant's current balance, creating one if it doesn't exist.
func (s *SettlementService) GetBalance(ctx context.Context, merchantID uuid.UUID) (*domain.MerchantBalance, error) {
	balance, err := s.balances.GetByMerchantID(ctx, merchantID)
	if err != nil {
		// Create a zero balance if none exists
		balance = domain.NewMerchantBalance(merchantID)
		if createErr := s.balances.Create(ctx, balance); createErr != nil {
			return nil, fmt.Errorf("creating balance: %w", createErr)
		}
	}
	return balance, nil
}

// CreditPayment credits a merchant's balance when a payment is confirmed.
func (s *SettlementService) CreditPayment(ctx context.Context, merchantID uuid.UUID, netUSDT, feesUSDT decimal.Decimal) error {
	balance, err := s.GetBalance(ctx, merchantID)
	if err != nil {
		return err
	}

	balance.Credit(netUSDT, feesUSDT)
	return s.balances.Update(ctx, balance)
}

// RequestWithdrawal creates a new withdrawal request and holds the balance.
func (s *SettlementService) RequestWithdrawal(ctx context.Context, merchantID uuid.UUID, amountUSDT, exchangeRate decimal.Decimal, bankName, bankAccountNo, bankAccountName string) (*domain.Withdrawal, error) {
	balance, err := s.GetBalance(ctx, merchantID)
	if err != nil {
		return nil, err
	}

	if err := balance.Hold(amountUSDT); err != nil {
		return nil, err
	}

	withdrawal, err := domain.NewWithdrawal(merchantID, amountUSDT, exchangeRate, bankName, bankAccountNo, bankAccountName)
	if err != nil {
		return nil, err
	}

	if err := s.withdrawals.Create(ctx, withdrawal); err != nil {
		return nil, fmt.Errorf("storing withdrawal: %w", err)
	}

	if err := s.balances.Update(ctx, balance); err != nil {
		return nil, fmt.Errorf("updating balance: %w", err)
	}

	return withdrawal, nil
}

// ApproveWithdrawal approves a pending withdrawal.
func (s *SettlementService) ApproveWithdrawal(ctx context.Context, withdrawalID, adminID uuid.UUID) error {
	withdrawal, err := s.withdrawals.GetByID(ctx, withdrawalID)
	if err != nil {
		return err
	}

	if err := withdrawal.Approve(adminID); err != nil {
		return err
	}

	return s.withdrawals.Update(ctx, withdrawal)
}

// RejectWithdrawal rejects a pending withdrawal and releases the held balance.
func (s *SettlementService) RejectWithdrawal(ctx context.Context, withdrawalID uuid.UUID, reason string) error {
	withdrawal, err := s.withdrawals.GetByID(ctx, withdrawalID)
	if err != nil {
		return err
	}

	balance, err := s.balances.GetByMerchantID(ctx, withdrawal.MerchantID)
	if err != nil {
		return err
	}

	if err := withdrawal.Reject(reason); err != nil {
		return err
	}

	balance.Release(withdrawal.AmountUSDT)

	if err := s.withdrawals.Update(ctx, withdrawal); err != nil {
		return err
	}

	return s.balances.Update(ctx, balance)
}

// CompleteWithdrawal marks a withdrawal as completed with a bank reference.
func (s *SettlementService) CompleteWithdrawal(ctx context.Context, withdrawalID uuid.UUID, bankReference string) error {
	withdrawal, err := s.withdrawals.GetByID(ctx, withdrawalID)
	if err != nil {
		return err
	}

	balance, err := s.balances.GetByMerchantID(ctx, withdrawal.MerchantID)
	if err != nil {
		return err
	}

	if err := withdrawal.Complete(bankReference); err != nil {
		return err
	}

	balance.Settle(withdrawal.AmountUSDT)

	if err := s.withdrawals.Update(ctx, withdrawal); err != nil {
		return err
	}

	return s.balances.Update(ctx, balance)
}

// ListWithdrawals returns withdrawals for a merchant.
func (s *SettlementService) ListWithdrawals(ctx context.Context, merchantID uuid.UUID, status *domain.WithdrawalStatus) ([]*domain.Withdrawal, error) {
	return s.withdrawals.ListByMerchant(ctx, merchantID, status)
}

// GetWithdrawal returns a single withdrawal by ID.
func (s *SettlementService) GetWithdrawal(ctx context.Context, id uuid.UUID) (*domain.Withdrawal, error) {
	return s.withdrawals.GetByID(ctx, id)
}

// RequestRefund creates a new refund request and debits the merchant balance.
func (s *SettlementService) RequestRefund(ctx context.Context, merchantID, paymentID uuid.UUID, paymentNo string, amountUSDT decimal.Decimal, reason string) (*domain.Refund, error) {
	balance, err := s.GetBalance(ctx, merchantID)
	if err != nil {
		return nil, err
	}

	if err := balance.Debit(amountUSDT); err != nil {
		return nil, err
	}

	refund, err := domain.NewRefund(merchantID, paymentID, paymentNo, amountUSDT, reason)
	if err != nil {
		return nil, err
	}

	if err := s.refunds.Create(ctx, refund); err != nil {
		return nil, fmt.Errorf("storing refund: %w", err)
	}

	if err := s.balances.Update(ctx, balance); err != nil {
		return nil, fmt.Errorf("updating balance: %w", err)
	}

	return refund, nil
}

// ListRefunds returns refunds for a merchant.
func (s *SettlementService) ListRefunds(ctx context.Context, merchantID uuid.UUID) ([]*domain.Refund, error) {
	return s.refunds.ListByMerchant(ctx, merchantID)
}

// GetRefund returns a single refund by ID.
func (s *SettlementService) GetRefund(ctx context.Context, id uuid.UUID) (*domain.Refund, error) {
	return s.refunds.GetByID(ctx, id)
}

// ApproveRefund approves a pending refund.
func (s *SettlementService) ApproveRefund(ctx context.Context, refundID, adminID uuid.UUID) error {
	refund, err := s.refunds.GetByID(ctx, refundID)
	if err != nil {
		return err
	}
	if err := refund.Approve(adminID); err != nil {
		return err
	}
	// Auto-complete after approval (crypto refunds are instant)
	if err := refund.Complete(); err != nil {
		return err
	}
	return s.refunds.Update(ctx, refund)
}

// RejectRefund rejects a pending refund and credits back the merchant balance.
func (s *SettlementService) RejectRefund(ctx context.Context, refundID uuid.UUID, reason string) error {
	refund, err := s.refunds.GetByID(ctx, refundID)
	if err != nil {
		return err
	}

	balance, err := s.balances.GetByMerchantID(ctx, refund.MerchantID)
	if err != nil {
		return err
	}

	if err := refund.Reject(reason); err != nil {
		return err
	}

	// Credit back the refund amount
	balance.Credit(refund.AmountUSDT, decimal.Zero)

	if err := s.refunds.Update(ctx, refund); err != nil {
		return err
	}
	return s.balances.Update(ctx, balance)
}
