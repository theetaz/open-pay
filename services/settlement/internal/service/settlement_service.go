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

// SettlementService orchestrates settlement and withdrawal operations.
type SettlementService struct {
	balances    BalanceRepository
	withdrawals WithdrawalRepository
}

// NewSettlementService creates a new SettlementService.
func NewSettlementService(balances BalanceRepository, withdrawals WithdrawalRepository) *SettlementService {
	return &SettlementService{balances: balances, withdrawals: withdrawals}
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
