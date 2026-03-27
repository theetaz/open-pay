package domain_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/openlankapay/openlankapay/services/payment/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDispute(t *testing.T) {
	paymentID := uuid.New()
	merchantID := uuid.New()

	t.Run("valid dispute", func(t *testing.T) {
		d, err := domain.NewDispute(paymentID, merchantID, "customer@example.com", "Item not received")
		require.NoError(t, err)
		assert.Equal(t, paymentID, d.PaymentID)
		assert.Equal(t, merchantID, d.MerchantID)
		assert.Equal(t, "customer@example.com", d.CustomerEmail)
		assert.Equal(t, "Item not received", d.Reason)
		assert.Equal(t, domain.DisputeOpened, d.Status)
		assert.NotEqual(t, uuid.Nil, d.ID)
		assert.False(t, d.CreatedAt.IsZero())
		assert.False(t, d.UpdatedAt.IsZero())
	})

	t.Run("missing payment ID", func(t *testing.T) {
		_, err := domain.NewDispute(uuid.Nil, merchantID, "customer@example.com", "Item not received")
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrInvalidDispute)
	})

	t.Run("missing merchant ID", func(t *testing.T) {
		_, err := domain.NewDispute(paymentID, uuid.Nil, "customer@example.com", "Item not received")
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrInvalidDispute)
	})

	t.Run("missing customer email", func(t *testing.T) {
		_, err := domain.NewDispute(paymentID, merchantID, "", "Item not received")
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrInvalidDispute)
	})

	t.Run("missing reason", func(t *testing.T) {
		_, err := domain.NewDispute(paymentID, merchantID, "customer@example.com", "")
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrInvalidDispute)
	})
}

func TestDisputeStatusTransitions(t *testing.T) {
	tests := []struct {
		name    string
		from    domain.DisputeStatus
		to      domain.DisputeStatus
		wantErr bool
	}{
		{"opened to merchant_response", domain.DisputeOpened, domain.DisputeMerchantResponse, false},
		{"merchant_response to under_review", domain.DisputeMerchantResponse, domain.DisputeUnderReview, false},
		{"under_review to resolved_customer", domain.DisputeUnderReview, domain.DisputeResolvedCustomer, false},
		{"under_review to resolved_merchant", domain.DisputeUnderReview, domain.DisputeResolvedMerchant, false},
		{"under_review to rejected", domain.DisputeUnderReview, domain.DisputeRejected, false},
		{"opened to under_review (invalid)", domain.DisputeOpened, domain.DisputeUnderReview, true},
		{"opened to resolved_customer (invalid)", domain.DisputeOpened, domain.DisputeResolvedCustomer, true},
		{"merchant_response to resolved_customer (invalid)", domain.DisputeMerchantResponse, domain.DisputeResolvedCustomer, true},
		{"resolved_customer to opened (invalid)", domain.DisputeResolvedCustomer, domain.DisputeOpened, true},
		{"resolved_merchant to opened (invalid)", domain.DisputeResolvedMerchant, domain.DisputeOpened, true},
		{"rejected to opened (invalid)", domain.DisputeRejected, domain.DisputeOpened, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &domain.Dispute{Status: tt.from}
			err := d.TransitionTo(tt.to)
			if tt.wantErr {
				require.Error(t, err)
				assert.ErrorIs(t, err, domain.ErrInvalidDisputeTransition)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.to, d.Status)
			}
		})
	}
}

func TestDisputeRespondAsMerchant(t *testing.T) {
	t.Run("valid response", func(t *testing.T) {
		d := &domain.Dispute{Status: domain.DisputeOpened}
		err := d.RespondAsMerchant("Item was shipped on time, tracking #12345")
		require.NoError(t, err)
		assert.Equal(t, domain.DisputeMerchantResponse, d.Status)
		assert.Equal(t, "Item was shipped on time, tracking #12345", d.MerchantResponse)
		assert.NotNil(t, d.MerchantRespondedAt)
	})

	t.Run("empty response", func(t *testing.T) {
		d := &domain.Dispute{Status: domain.DisputeOpened}
		err := d.RespondAsMerchant("")
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrInvalidDispute)
		assert.Equal(t, domain.DisputeOpened, d.Status)
	})

	t.Run("invalid state for response", func(t *testing.T) {
		d := &domain.Dispute{Status: domain.DisputeUnderReview}
		err := d.RespondAsMerchant("Some response")
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrInvalidDisputeTransition)
	})
}

func TestDisputeResolve(t *testing.T) {
	adminID := uuid.New()

	t.Run("resolve in favor of customer", func(t *testing.T) {
		d := &domain.Dispute{Status: domain.DisputeUnderReview}
		err := d.Resolve(adminID, "Customer provided valid evidence", true)
		require.NoError(t, err)
		assert.Equal(t, domain.DisputeResolvedCustomer, d.Status)
		assert.Equal(t, &adminID, d.ResolvedBy)
		assert.NotNil(t, d.ResolvedAt)
		assert.Equal(t, "Customer provided valid evidence", d.AdminNotes)
	})

	t.Run("resolve in favor of merchant", func(t *testing.T) {
		d := &domain.Dispute{Status: domain.DisputeUnderReview}
		err := d.Resolve(adminID, "Merchant provided delivery proof", false)
		require.NoError(t, err)
		assert.Equal(t, domain.DisputeResolvedMerchant, d.Status)
		assert.Equal(t, &adminID, d.ResolvedBy)
		assert.NotNil(t, d.ResolvedAt)
		assert.Equal(t, "Merchant provided delivery proof", d.AdminNotes)
	})

	t.Run("cannot resolve from opened state", func(t *testing.T) {
		d := &domain.Dispute{Status: domain.DisputeOpened}
		err := d.Resolve(adminID, "notes", true)
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrInvalidDisputeTransition)
	})
}

func TestDisputeReject(t *testing.T) {
	adminID := uuid.New()

	t.Run("reject from under_review", func(t *testing.T) {
		d := &domain.Dispute{Status: domain.DisputeUnderReview}
		err := d.Reject(adminID, "Fraudulent claim")
		require.NoError(t, err)
		assert.Equal(t, domain.DisputeRejected, d.Status)
		assert.Equal(t, &adminID, d.ResolvedBy)
		assert.NotNil(t, d.ResolvedAt)
		assert.Equal(t, "Fraudulent claim", d.AdminNotes)
	})

	t.Run("cannot reject from opened state", func(t *testing.T) {
		d := &domain.Dispute{Status: domain.DisputeOpened}
		err := d.Reject(adminID, "notes")
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrInvalidDisputeTransition)
	})

	t.Run("cannot reject already resolved dispute", func(t *testing.T) {
		d := &domain.Dispute{Status: domain.DisputeResolvedCustomer}
		err := d.Reject(adminID, "notes")
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrInvalidDisputeTransition)
	})
}
