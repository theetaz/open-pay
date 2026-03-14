package domain_test

import (
	"testing"

	"github.com/openlankapay/openlankapay/services/merchant/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMerchant(t *testing.T) {
	t.Run("valid merchant", func(t *testing.T) {
		m, err := domain.NewMerchant("Test Business", "test@example.com")
		require.NoError(t, err)
		assert.Equal(t, "Test Business", m.BusinessName)
		assert.Equal(t, "test@example.com", m.ContactEmail)
		assert.Equal(t, domain.KYCPending, m.KYCStatus)
		assert.Equal(t, domain.MerchantActive, m.Status)
		assert.NotEmpty(t, m.ID)
	})

	t.Run("empty business name", func(t *testing.T) {
		_, err := domain.NewMerchant("", "test@example.com")
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrInvalidMerchant)
	})

	t.Run("empty email", func(t *testing.T) {
		_, err := domain.NewMerchant("Test Business", "")
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrInvalidMerchant)
	})

	t.Run("invalid email format", func(t *testing.T) {
		_, err := domain.NewMerchant("Test Business", "not-an-email")
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrInvalidMerchant)
	})
}

func TestMerchantKYCTransitions(t *testing.T) {
	tests := []struct {
		name    string
		from    domain.KYCStatus
		to      domain.KYCStatus
		wantErr bool
	}{
		{"pending to under_review", domain.KYCPending, domain.KYCUnderReview, false},
		{"pending to instant_access", domain.KYCPending, domain.KYCInstantAccess, false},
		{"under_review to approved", domain.KYCUnderReview, domain.KYCApproved, false},
		{"under_review to rejected", domain.KYCUnderReview, domain.KYCRejected, false},
		{"instant_access to under_review", domain.KYCInstantAccess, domain.KYCUnderReview, false},
		{"instant_access to approved", domain.KYCInstantAccess, domain.KYCApproved, false},
		{"approved to rejected (invalid)", domain.KYCApproved, domain.KYCRejected, true},
		{"rejected to approved (invalid)", domain.KYCRejected, domain.KYCApproved, true},
		{"pending to approved (invalid)", domain.KYCPending, domain.KYCApproved, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &domain.Merchant{KYCStatus: tt.from}
			err := m.TransitionKYC(tt.to)
			if tt.wantErr {
				require.Error(t, err)
				assert.ErrorIs(t, err, domain.ErrInvalidKYCTransition)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.to, m.KYCStatus)
			}
		})
	}
}

func TestMerchantCanAcceptPayments(t *testing.T) {
	t.Run("approved merchant can accept", func(t *testing.T) {
		m := &domain.Merchant{KYCStatus: domain.KYCApproved, Status: domain.MerchantActive}
		assert.True(t, m.CanAcceptPayments())
	})

	t.Run("instant_access merchant can accept", func(t *testing.T) {
		m := &domain.Merchant{KYCStatus: domain.KYCInstantAccess, Status: domain.MerchantActive}
		assert.True(t, m.CanAcceptPayments())
	})

	t.Run("pending merchant cannot accept", func(t *testing.T) {
		m := &domain.Merchant{KYCStatus: domain.KYCPending, Status: domain.MerchantActive}
		assert.False(t, m.CanAcceptPayments())
	})

	t.Run("inactive merchant cannot accept", func(t *testing.T) {
		m := &domain.Merchant{KYCStatus: domain.KYCApproved, Status: domain.MerchantInactive}
		assert.False(t, m.CanAcceptPayments())
	})
}
