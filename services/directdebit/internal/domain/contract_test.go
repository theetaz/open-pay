package domain_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/openlankapay/openlankapay/services/directdebit/internal/domain"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewContract(t *testing.T) {
	merchantID := uuid.New()
	scenarioID := uuid.New()

	t.Run("valid contract", func(t *testing.T) {
		c, err := domain.NewContract(merchantID, "Monthly Service", scenarioID, domain.ProviderBinancePay, decimal.NewFromFloat(100), "https://example.com/return", "https://example.com/cancel")
		require.NoError(t, err)
		assert.Equal(t, "Monthly Service", c.ServiceName)
		assert.Equal(t, domain.ContractInitiated, c.Status)
		assert.Equal(t, domain.ProviderBinancePay, c.PaymentProvider)
		assert.Equal(t, "USDT", c.Currency)
		assert.NotEmpty(t, c.MerchantContractCode)
	})

	t.Run("empty service name", func(t *testing.T) {
		_, err := domain.NewContract(merchantID, "", scenarioID, domain.ProviderBinancePay, decimal.NewFromFloat(100), "https://example.com/return", "https://example.com/cancel")
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrInvalidContract)
	})

	t.Run("service name too long", func(t *testing.T) {
		longName := ""
		for i := 0; i < 65; i++ {
			longName += "x"
		}
		_, err := domain.NewContract(merchantID, longName, scenarioID, domain.ProviderBinancePay, decimal.NewFromFloat(100), "https://example.com/return", "https://example.com/cancel")
		require.Error(t, err)
	})

	t.Run("limit too low", func(t *testing.T) {
		_, err := domain.NewContract(merchantID, "Test", scenarioID, domain.ProviderBinancePay, decimal.NewFromFloat(0.001), "https://example.com/return", "https://example.com/cancel")
		require.Error(t, err)
	})

	t.Run("missing return URL", func(t *testing.T) {
		_, err := domain.NewContract(merchantID, "Test", scenarioID, domain.ProviderBinancePay, decimal.NewFromFloat(100), "", "https://example.com/cancel")
		require.Error(t, err)
	})

	t.Run("missing cancel URL", func(t *testing.T) {
		_, err := domain.NewContract(merchantID, "Test", scenarioID, domain.ProviderBinancePay, decimal.NewFromFloat(100), "https://example.com/return", "")
		require.Error(t, err)
	})
}

func TestContractStatusTransitions(t *testing.T) {
	tests := []struct {
		name    string
		from    domain.ContractStatus
		to      domain.ContractStatus
		wantErr bool
	}{
		{"initiated to signed", domain.ContractInitiated, domain.ContractSigned, false},
		{"initiated to expired", domain.ContractInitiated, domain.ContractExpired, false},
		{"signed to terminated", domain.ContractSigned, domain.ContractTerminated, false},
		{"signed to expired", domain.ContractSigned, domain.ContractExpired, false},
		{"terminated to signed (invalid)", domain.ContractTerminated, domain.ContractSigned, true},
		{"expired to signed (invalid)", domain.ContractExpired, domain.ContractSigned, true},
		{"initiated to terminated (invalid)", domain.ContractInitiated, domain.ContractTerminated, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &domain.Contract{Status: tt.from}
			err := c.TransitionTo(tt.to)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.to, c.Status)
			}
		})
	}
}

func TestContractTerminate(t *testing.T) {
	merchantID := uuid.New()
	scenarioID := uuid.New()

	t.Run("terminate signed contract", func(t *testing.T) {
		c, _ := domain.NewContract(merchantID, "Test", scenarioID, domain.ProviderBinancePay, decimal.NewFromFloat(100), "https://example.com/return", "https://example.com/cancel")
		_ = c.TransitionTo(domain.ContractSigned)

		err := c.Terminate("No longer needed")
		require.NoError(t, err)
		assert.Equal(t, domain.ContractTerminated, c.Status)
		assert.Equal(t, "No longer needed", c.TerminationNotes)
		assert.NotNil(t, c.TerminationTime)
		assert.NotNil(t, c.TerminationWay)
	})

	t.Run("terminate initiated contract fails", func(t *testing.T) {
		c, _ := domain.NewContract(merchantID, "Test", scenarioID, domain.ProviderBinancePay, decimal.NewFromFloat(100), "https://example.com/return", "https://example.com/cancel")

		err := c.Terminate("Should fail")
		require.Error(t, err)
	})
}

func TestCanExecutePayment(t *testing.T) {
	merchantID := uuid.New()
	scenarioID := uuid.New()

	t.Run("signed contract within limit", func(t *testing.T) {
		c, _ := domain.NewContract(merchantID, "Test", scenarioID, domain.ProviderBinancePay, decimal.NewFromFloat(100), "https://example.com/return", "https://example.com/cancel")
		_ = c.TransitionTo(domain.ContractSigned)

		err := c.CanExecutePayment(decimal.NewFromFloat(50))
		require.NoError(t, err)
	})

	t.Run("amount exceeds limit", func(t *testing.T) {
		c, _ := domain.NewContract(merchantID, "Test", scenarioID, domain.ProviderBinancePay, decimal.NewFromFloat(100), "https://example.com/return", "https://example.com/cancel")
		_ = c.TransitionTo(domain.ContractSigned)

		err := c.CanExecutePayment(decimal.NewFromFloat(150))
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrAmountExceedsLimit)
	})

	t.Run("contract not signed", func(t *testing.T) {
		c, _ := domain.NewContract(merchantID, "Test", scenarioID, domain.ProviderBinancePay, decimal.NewFromFloat(100), "https://example.com/return", "https://example.com/cancel")

		err := c.CanExecutePayment(decimal.NewFromFloat(50))
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrContractNotSigned)
	})
}

func TestNewPayment(t *testing.T) {
	merchantID := uuid.New()
	scenarioID := uuid.New()

	t.Run("valid payment", func(t *testing.T) {
		c, _ := domain.NewContract(merchantID, "Test", scenarioID, domain.ProviderBinancePay, decimal.NewFromFloat(100), "https://example.com/return", "https://example.com/cancel")
		_ = c.TransitionTo(domain.ContractSigned)

		p, err := domain.NewPayment(c, decimal.NewFromFloat(25), "Monthly Charge")
		require.NoError(t, err)
		assert.Equal(t, domain.PaymentInitiated, p.Status)
		assert.Equal(t, "Monthly Charge", p.ProductName)
		assert.NotEmpty(t, p.PaymentNo)
		assert.True(t, p.NetAmountUSDT.LessThan(p.GrossAmountUSDT))
		assert.True(t, p.TotalFeesUSDT.GreaterThan(decimal.Zero))
	})

	t.Run("payment exceeds limit", func(t *testing.T) {
		c, _ := domain.NewContract(merchantID, "Test", scenarioID, domain.ProviderBinancePay, decimal.NewFromFloat(100), "https://example.com/return", "https://example.com/cancel")
		_ = c.TransitionTo(domain.ContractSigned)

		_, err := domain.NewPayment(c, decimal.NewFromFloat(200), "Too Much")
		require.Error(t, err)
	})

	t.Run("empty product name", func(t *testing.T) {
		c, _ := domain.NewContract(merchantID, "Test", scenarioID, domain.ProviderBinancePay, decimal.NewFromFloat(100), "https://example.com/return", "https://example.com/cancel")
		_ = c.TransitionTo(domain.ContractSigned)

		_, err := domain.NewPayment(c, decimal.NewFromFloat(25), "")
		require.Error(t, err)
	})

	t.Run("contract not signed", func(t *testing.T) {
		c, _ := domain.NewContract(merchantID, "Test", scenarioID, domain.ProviderBinancePay, decimal.NewFromFloat(100), "https://example.com/return", "https://example.com/cancel")

		_, err := domain.NewPayment(c, decimal.NewFromFloat(25), "Test")
		require.Error(t, err)
	})
}

func TestPaymentMarkPaid(t *testing.T) {
	merchantID := uuid.New()
	scenarioID := uuid.New()
	c, _ := domain.NewContract(merchantID, "Test", scenarioID, domain.ProviderBinancePay, decimal.NewFromFloat(100), "https://example.com/return", "https://example.com/cancel")
	_ = c.TransitionTo(domain.ContractSigned)
	p, _ := domain.NewPayment(c, decimal.NewFromFloat(25), "Charge")

	p.MarkPaid("provider-pay-123")
	assert.Equal(t, domain.PaymentPaid, p.Status)
	assert.Equal(t, "provider-pay-123", p.PayID)
}

func TestPaymentMarkFailed(t *testing.T) {
	merchantID := uuid.New()
	scenarioID := uuid.New()
	c, _ := domain.NewContract(merchantID, "Test", scenarioID, domain.ProviderBinancePay, decimal.NewFromFloat(100), "https://example.com/return", "https://example.com/cancel")
	_ = c.TransitionTo(domain.ContractSigned)
	p, _ := domain.NewPayment(c, decimal.NewFromFloat(25), "Charge")

	p.MarkFailed()
	assert.Equal(t, domain.PaymentFailed, p.Status)
}
