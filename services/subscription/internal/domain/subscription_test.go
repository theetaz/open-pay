package domain_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/openlankapay/openlankapay/services/subscription/internal/domain"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSubscriptionPlan(t *testing.T) {
	merchantID := uuid.New()

	t.Run("valid monthly plan", func(t *testing.T) {
		plan, err := domain.NewSubscriptionPlan(merchantID, "Premium", decimal.NewFromFloat(25), "USDT", domain.IntervalMonthly, 1)
		require.NoError(t, err)
		assert.Equal(t, "Premium", plan.Name)
		assert.Equal(t, domain.IntervalMonthly, plan.IntervalType)
		assert.Equal(t, 1, plan.IntervalCount)
		assert.Equal(t, domain.PlanActive, plan.Status)
	})

	t.Run("valid yearly plan", func(t *testing.T) {
		plan, err := domain.NewSubscriptionPlan(merchantID, "Annual", decimal.NewFromFloat(250), "USDT", domain.IntervalYearly, 1)
		require.NoError(t, err)
		assert.Equal(t, domain.IntervalYearly, plan.IntervalType)
	})

	t.Run("zero amount invalid", func(t *testing.T) {
		_, err := domain.NewSubscriptionPlan(merchantID, "Bad", decimal.Zero, "USDT", domain.IntervalMonthly, 1)
		require.Error(t, err)
	})

	t.Run("invalid interval", func(t *testing.T) {
		_, err := domain.NewSubscriptionPlan(merchantID, "Bad", decimal.NewFromFloat(10), "USDT", "BIWEEKLY", 1)
		require.Error(t, err)
	})
}

func TestNewSubscription(t *testing.T) {
	planID := uuid.New()
	merchantID := uuid.New()

	t.Run("valid subscription", func(t *testing.T) {
		sub, err := domain.NewSubscription(planID, merchantID, "user@test.com", decimal.NewFromFloat(25), domain.IntervalMonthly, 1, 0)
		require.NoError(t, err)
		assert.Equal(t, domain.SubActive, sub.Status)
		assert.Equal(t, "user@test.com", sub.SubscriberEmail)
		assert.False(t, sub.NextBillingDate.IsZero())
	})

	t.Run("with trial days", func(t *testing.T) {
		sub, err := domain.NewSubscription(planID, merchantID, "trial@test.com", decimal.NewFromFloat(25), domain.IntervalMonthly, 1, 14)
		require.NoError(t, err)
		assert.Equal(t, domain.SubTrial, sub.Status)
		assert.NotNil(t, sub.TrialEnd)
		assert.WithinDuration(t, time.Now().Add(14*24*time.Hour), *sub.TrialEnd, time.Minute)
	})

	t.Run("empty email invalid", func(t *testing.T) {
		_, err := domain.NewSubscription(planID, merchantID, "", decimal.NewFromFloat(25), domain.IntervalMonthly, 1, 0)
		require.Error(t, err)
	})
}

func TestSubscriptionStatusTransitions(t *testing.T) {
	tests := []struct {
		name    string
		from    domain.SubscriptionStatus
		to      domain.SubscriptionStatus
		wantErr bool
	}{
		{"trial to active", domain.SubTrial, domain.SubActive, false},
		{"active to past_due", domain.SubActive, domain.SubPastDue, false},
		{"active to paused", domain.SubActive, domain.SubPaused, false},
		{"active to cancelled", domain.SubActive, domain.SubCancelled, false},
		{"past_due to active", domain.SubPastDue, domain.SubActive, false},
		{"past_due to cancelled", domain.SubPastDue, domain.SubCancelled, false},
		{"paused to active", domain.SubPaused, domain.SubActive, false},
		{"cancelled to active (invalid)", domain.SubCancelled, domain.SubActive, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sub := &domain.Subscription{Status: tt.from}
			err := sub.TransitionTo(tt.to)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.to, sub.Status)
			}
		})
	}
}

func TestSubscriptionCancel(t *testing.T) {
	planID := uuid.New()
	merchantID := uuid.New()
	sub, _ := domain.NewSubscription(planID, merchantID, "cancel@test.com", decimal.NewFromFloat(25), domain.IntervalMonthly, 1, 0)

	t.Run("cancel at period end", func(t *testing.T) {
		sub.CancelAtPeriodEnd("No longer needed")
		assert.True(t, sub.CancelAtEnd)
		assert.NotNil(t, sub.CancelledAt)
		assert.Equal(t, "No longer needed", sub.CancellationReason)
		// Still active until period ends
		assert.Equal(t, domain.SubActive, sub.Status)
	})
}

func TestCalculateNextBillingDate(t *testing.T) {
	now := time.Now()

	t.Run("monthly", func(t *testing.T) {
		next := domain.CalculateNextBillingDate(now, domain.IntervalMonthly, 1)
		assert.WithinDuration(t, now.AddDate(0, 1, 0), next, time.Minute)
	})

	t.Run("weekly", func(t *testing.T) {
		next := domain.CalculateNextBillingDate(now, domain.IntervalWeekly, 1)
		assert.WithinDuration(t, now.Add(7*24*time.Hour), next, time.Minute)
	})

	t.Run("every 3 months", func(t *testing.T) {
		next := domain.CalculateNextBillingDate(now, domain.IntervalMonthly, 3)
		assert.WithinDuration(t, now.AddDate(0, 3, 0), next, time.Minute)
	})

	t.Run("yearly", func(t *testing.T) {
		next := domain.CalculateNextBillingDate(now, domain.IntervalYearly, 1)
		assert.WithinDuration(t, now.AddDate(1, 0, 0), next, time.Minute)
	})
}
