package domain_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/openlankapay/openlankapay/services/settlement/internal/domain"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMerchantBalance(t *testing.T) {
	merchantID := uuid.New()

	t.Run("new balance starts at zero", func(t *testing.T) {
		b := domain.NewMerchantBalance(merchantID)
		assert.True(t, b.AvailableUSDT.IsZero())
		assert.True(t, b.PendingUSDT.IsZero())
		assert.True(t, b.TotalEarnedUSDT.IsZero())
	})

	t.Run("credit increases available and total earned", func(t *testing.T) {
		b := domain.NewMerchantBalance(merchantID)
		b.Credit(decimal.NewFromFloat(9.80), decimal.NewFromFloat(0.20))

		assert.True(t, decimal.NewFromFloat(9.80).Equal(b.AvailableUSDT))
		assert.True(t, decimal.NewFromFloat(9.80).Equal(b.TotalEarnedUSDT))
		assert.True(t, decimal.NewFromFloat(0.20).Equal(b.TotalFeesUSDT))
	})

	t.Run("multiple credits accumulate", func(t *testing.T) {
		b := domain.NewMerchantBalance(merchantID)
		b.Credit(decimal.NewFromFloat(9.80), decimal.NewFromFloat(0.20))
		b.Credit(decimal.NewFromFloat(19.60), decimal.NewFromFloat(0.40))

		assert.True(t, decimal.NewFromFloat(29.40).Equal(b.AvailableUSDT))
		assert.True(t, decimal.NewFromFloat(0.60).Equal(b.TotalFeesUSDT))
	})

	t.Run("hold moves from available to pending", func(t *testing.T) {
		b := domain.NewMerchantBalance(merchantID)
		b.Credit(decimal.NewFromFloat(100), decimal.Zero)

		err := b.Hold(decimal.NewFromFloat(50))
		require.NoError(t, err)
		assert.True(t, decimal.NewFromFloat(50).Equal(b.AvailableUSDT))
		assert.True(t, decimal.NewFromFloat(50).Equal(b.PendingUSDT))
	})

	t.Run("hold exceeding available fails", func(t *testing.T) {
		b := domain.NewMerchantBalance(merchantID)
		b.Credit(decimal.NewFromFloat(10), decimal.Zero)

		err := b.Hold(decimal.NewFromFloat(20))
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrInsufficientBalance)
	})

	t.Run("release moves from pending back to available", func(t *testing.T) {
		b := domain.NewMerchantBalance(merchantID)
		b.Credit(decimal.NewFromFloat(100), decimal.Zero)
		_ = b.Hold(decimal.NewFromFloat(50))

		b.Release(decimal.NewFromFloat(50))
		assert.True(t, decimal.NewFromFloat(100).Equal(b.AvailableUSDT))
		assert.True(t, b.PendingUSDT.IsZero())
	})

	t.Run("settle deducts from pending and tracks withdrawn", func(t *testing.T) {
		b := domain.NewMerchantBalance(merchantID)
		b.Credit(decimal.NewFromFloat(100), decimal.Zero)
		_ = b.Hold(decimal.NewFromFloat(100))

		b.Settle(decimal.NewFromFloat(100))
		assert.True(t, b.PendingUSDT.IsZero())
		assert.True(t, decimal.NewFromFloat(100).Equal(b.TotalWithdrawnUSDT))
	})
}

func TestNewWithdrawal(t *testing.T) {
	merchantID := uuid.New()

	t.Run("valid withdrawal", func(t *testing.T) {
		w, err := domain.NewWithdrawal(merchantID, decimal.NewFromFloat(100), decimal.NewFromFloat(325), "BOC", "1234567890", "John Doe")
		require.NoError(t, err)
		assert.Equal(t, domain.WithdrawalRequested, w.Status)
		assert.True(t, decimal.NewFromFloat(100).Equal(w.AmountUSDT))
		assert.True(t, decimal.NewFromFloat(32500).Equal(w.AmountLKR))
	})

	t.Run("zero amount", func(t *testing.T) {
		_, err := domain.NewWithdrawal(merchantID, decimal.Zero, decimal.NewFromFloat(325), "BOC", "123", "John")
		require.Error(t, err)
	})
}

func TestWithdrawalLifecycle(t *testing.T) {
	merchantID := uuid.New()
	w, _ := domain.NewWithdrawal(merchantID, decimal.NewFromFloat(100), decimal.NewFromFloat(325), "BOC", "123", "John")

	t.Run("approve", func(t *testing.T) {
		adminID := uuid.New()
		err := w.Approve(adminID)
		require.NoError(t, err)
		assert.Equal(t, domain.WithdrawalApproved, w.Status)
		assert.Equal(t, &adminID, w.ApprovedBy)
	})

	t.Run("complete", func(t *testing.T) {
		err := w.Complete("BANK-REF-001")
		require.NoError(t, err)
		assert.Equal(t, domain.WithdrawalCompleted, w.Status)
		assert.Equal(t, "BANK-REF-001", w.BankReference)
	})

	t.Run("cannot approve completed", func(t *testing.T) {
		err := w.Approve(uuid.New())
		require.Error(t, err)
	})
}

func TestWithdrawalReject(t *testing.T) {
	merchantID := uuid.New()
	w, _ := domain.NewWithdrawal(merchantID, decimal.NewFromFloat(100), decimal.NewFromFloat(325), "BOC", "123", "John")

	err := w.Reject("Suspicious activity")
	require.NoError(t, err)
	assert.Equal(t, domain.WithdrawalRejected, w.Status)
	assert.Equal(t, "Suspicious activity", w.RejectedReason)
}

func TestBalanceDebit(t *testing.T) {
	merchantID := uuid.New()

	t.Run("debit within available", func(t *testing.T) {
		b := domain.NewMerchantBalance(merchantID)
		b.Credit(decimal.NewFromFloat(100), decimal.Zero)
		err := b.Debit(decimal.NewFromFloat(25))
		require.NoError(t, err)
		assert.True(t, decimal.NewFromFloat(75).Equal(b.AvailableUSDT))
	})

	t.Run("debit exceeding available fails", func(t *testing.T) {
		b := domain.NewMerchantBalance(merchantID)
		b.Credit(decimal.NewFromFloat(10), decimal.Zero)
		err := b.Debit(decimal.NewFromFloat(20))
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrInsufficientBalance)
	})
}

func TestNewRefund(t *testing.T) {
	merchantID := uuid.New()
	paymentID := uuid.New()

	t.Run("valid refund", func(t *testing.T) {
		r, err := domain.NewRefund(merchantID, paymentID, "PAY-001", decimal.NewFromFloat(25), "Customer requested")
		require.NoError(t, err)
		assert.Equal(t, domain.RefundPending, r.Status)
		assert.True(t, decimal.NewFromFloat(25).Equal(r.AmountUSDT))
		assert.Equal(t, "Customer requested", r.Reason)
	})

	t.Run("zero amount", func(t *testing.T) {
		_, err := domain.NewRefund(merchantID, paymentID, "PAY-001", decimal.Zero, "reason")
		require.Error(t, err)
	})

	t.Run("empty reason", func(t *testing.T) {
		_, err := domain.NewRefund(merchantID, paymentID, "PAY-001", decimal.NewFromFloat(25), "")
		require.Error(t, err)
	})

	t.Run("empty payment number", func(t *testing.T) {
		_, err := domain.NewRefund(merchantID, paymentID, "", decimal.NewFromFloat(25), "reason")
		require.Error(t, err)
	})
}

func TestRefundLifecycle(t *testing.T) {
	merchantID := uuid.New()
	paymentID := uuid.New()

	t.Run("approve and complete", func(t *testing.T) {
		r, _ := domain.NewRefund(merchantID, paymentID, "PAY-001", decimal.NewFromFloat(25), "Duplicate charge")
		adminID := uuid.New()

		err := r.Approve(adminID)
		require.NoError(t, err)
		assert.Equal(t, domain.RefundApproved, r.Status)
		assert.Equal(t, &adminID, r.ApprovedBy)

		err = r.Complete()
		require.NoError(t, err)
		assert.Equal(t, domain.RefundCompleted, r.Status)
		assert.NotNil(t, r.CompletedAt)
	})

	t.Run("reject", func(t *testing.T) {
		r, _ := domain.NewRefund(merchantID, paymentID, "PAY-002", decimal.NewFromFloat(25), "reason")
		err := r.Reject("Not eligible")
		require.NoError(t, err)
		assert.Equal(t, domain.RefundRejected, r.Status)
	})

	t.Run("cannot approve non-pending", func(t *testing.T) {
		r, _ := domain.NewRefund(merchantID, paymentID, "PAY-003", decimal.NewFromFloat(25), "reason")
		_ = r.Reject("no")
		err := r.Approve(uuid.New())
		require.Error(t, err)
	})

	t.Run("cannot complete non-approved", func(t *testing.T) {
		r, _ := domain.NewRefund(merchantID, paymentID, "PAY-004", decimal.NewFromFloat(25), "reason")
		err := r.Complete()
		require.Error(t, err)
	})
}
