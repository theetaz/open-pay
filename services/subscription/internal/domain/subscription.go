package domain

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

var (
	ErrInvalidPlan               = errors.New("invalid subscription plan")
	ErrInvalidSubscription       = errors.New("invalid subscription")
	ErrInvalidStatusTransition   = errors.New("invalid subscription status transition")
	ErrPlanNotFound              = errors.New("subscription plan not found")
	ErrSubscriptionNotFound      = errors.New("subscription not found")
)

// Interval types for billing cycles.
type IntervalType = string

const (
	IntervalDaily   IntervalType = "DAILY"
	IntervalWeekly  IntervalType = "WEEKLY"
	IntervalMonthly IntervalType = "MONTHLY"
	IntervalYearly  IntervalType = "YEARLY"
)

var validIntervals = map[string]bool{
	IntervalDaily: true, IntervalWeekly: true,
	IntervalMonthly: true, IntervalYearly: true,
}

// Plan status values.
type PlanStatus = string

const (
	PlanActive   PlanStatus = "ACTIVE"
	PlanArchived PlanStatus = "ARCHIVED"
)

// Subscription status values.
type SubscriptionStatus = string

const (
	SubTrial     SubscriptionStatus = "TRIAL"
	SubActive    SubscriptionStatus = "ACTIVE"
	SubPastDue   SubscriptionStatus = "PAST_DUE"
	SubPaused    SubscriptionStatus = "PAUSED"
	SubCancelled SubscriptionStatus = "CANCELLED"
)

var validSubTransitions = map[SubscriptionStatus][]SubscriptionStatus{
	SubTrial:   {SubActive, SubCancelled},
	SubActive:  {SubPastDue, SubPaused, SubCancelled},
	SubPastDue: {SubActive, SubCancelled},
	SubPaused:  {SubActive, SubCancelled},
}

// SubscriptionPlan defines a recurring payment configuration.
type SubscriptionPlan struct {
	ID            uuid.UUID
	MerchantID    uuid.UUID
	Name          string
	Description   string
	Amount        decimal.Decimal
	Currency      string
	IntervalType  IntervalType
	IntervalCount int
	TrialDays     int
	MaxSubscribers *int
	ContractAddress string
	Status        PlanStatus
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     *time.Time
}

// NewSubscriptionPlan creates a validated subscription plan.
func NewSubscriptionPlan(merchantID uuid.UUID, name string, amount decimal.Decimal, currency string, interval IntervalType, intervalCount int) (*SubscriptionPlan, error) {
	if name == "" {
		return nil, fmt.Errorf("%w: name is required", ErrInvalidPlan)
	}
	if amount.LessThanOrEqual(decimal.Zero) {
		return nil, fmt.Errorf("%w: amount must be positive", ErrInvalidPlan)
	}
	if !validIntervals[interval] {
		return nil, fmt.Errorf("%w: unsupported interval %s", ErrInvalidPlan, interval)
	}
	if intervalCount < 1 {
		return nil, fmt.Errorf("%w: interval count must be at least 1", ErrInvalidPlan)
	}

	now := time.Now().UTC()
	return &SubscriptionPlan{
		ID:            uuid.New(),
		MerchantID:    merchantID,
		Name:          name,
		Amount:        amount,
		Currency:      currency,
		IntervalType:  interval,
		IntervalCount: intervalCount,
		Status:        PlanActive,
		CreatedAt:     now,
		UpdatedAt:     now,
	}, nil
}

// Subscription represents an active subscription for a customer.
type Subscription struct {
	ID                  uuid.UUID
	PlanID              uuid.UUID
	MerchantID          uuid.UUID
	SubscriberEmail     string
	SubscriberWallet    string
	Status              SubscriptionStatus
	CurrentPeriodStart  time.Time
	CurrentPeriodEnd    time.Time
	NextBillingDate     time.Time
	TrialEnd            *time.Time
	CancelAtEnd         bool
	CancelledAt         *time.Time
	CancellationReason  string
	TotalPaidUSDT       decimal.Decimal
	BillingCount        int
	FailedPaymentCount  int
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

// NewSubscription creates a new subscription, optionally with a trial period.
func NewSubscription(planID, merchantID uuid.UUID, email string, amount decimal.Decimal, interval IntervalType, intervalCount, trialDays int) (*Subscription, error) {
	if email == "" {
		return nil, fmt.Errorf("%w: subscriber email is required", ErrInvalidSubscription)
	}

	now := time.Now().UTC()
	status := SubActive
	var trialEnd *time.Time

	if trialDays > 0 {
		status = SubTrial
		t := now.Add(time.Duration(trialDays) * 24 * time.Hour)
		trialEnd = &t
	}

	nextBilling := CalculateNextBillingDate(now, interval, intervalCount)

	return &Subscription{
		ID:                 uuid.New(),
		PlanID:             planID,
		MerchantID:         merchantID,
		SubscriberEmail:    email,
		Status:             status,
		CurrentPeriodStart: now,
		CurrentPeriodEnd:   nextBilling,
		NextBillingDate:    nextBilling,
		TrialEnd:           trialEnd,
		CreatedAt:          now,
		UpdatedAt:          now,
	}, nil
}

// TransitionTo moves the subscription to a new status if valid.
func (s *Subscription) TransitionTo(to SubscriptionStatus) error {
	allowed, ok := validSubTransitions[s.Status]
	if !ok {
		return fmt.Errorf("%w: no transitions from %s", ErrInvalidStatusTransition, s.Status)
	}
	for _, a := range allowed {
		if a == to {
			s.Status = to
			s.UpdatedAt = time.Now().UTC()
			return nil
		}
	}
	return fmt.Errorf("%w: cannot transition from %s to %s", ErrInvalidStatusTransition, s.Status, to)
}

// CancelAtPeriodEnd marks the subscription to cancel at the end of the current period.
func (s *Subscription) CancelAtPeriodEnd(reason string) {
	now := time.Now().UTC()
	s.CancelAtEnd = true
	s.CancelledAt = &now
	s.CancellationReason = reason
	s.UpdatedAt = now
}

// CalculateNextBillingDate computes the next billing date from a reference time.
func CalculateNextBillingDate(from time.Time, interval IntervalType, count int) time.Time {
	switch interval {
	case IntervalDaily:
		return from.AddDate(0, 0, count)
	case IntervalWeekly:
		return from.AddDate(0, 0, 7*count)
	case IntervalMonthly:
		return from.AddDate(0, count, 0)
	case IntervalYearly:
		return from.AddDate(count, 0, 0)
	default:
		return from.AddDate(0, count, 0)
	}
}
