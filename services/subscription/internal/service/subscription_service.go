package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/openlankapay/openlankapay/services/subscription/internal/domain"
)

type PlanRepository interface {
	Create(ctx context.Context, plan *domain.SubscriptionPlan) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.SubscriptionPlan, error)
	ListByMerchant(ctx context.Context, merchantID uuid.UUID) ([]*domain.SubscriptionPlan, error)
	Update(ctx context.Context, plan *domain.SubscriptionPlan) error
}

type SubscriptionRepository interface {
	Create(ctx context.Context, sub *domain.Subscription) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Subscription, error)
	ListByMerchant(ctx context.Context, merchantID uuid.UUID) ([]*domain.Subscription, error)
	Update(ctx context.Context, sub *domain.Subscription) error
}

type SubscriptionService struct {
	plans PlanRepository
	subs  SubscriptionRepository
}

func NewSubscriptionService(plans PlanRepository, subs SubscriptionRepository) *SubscriptionService {
	return &SubscriptionService{plans: plans, subs: subs}
}

// CreatePlan creates a new subscription plan.
func (s *SubscriptionService) CreatePlan(ctx context.Context, merchantID uuid.UUID, name, description string, amount decimal.Decimal, currency string, interval domain.IntervalType, intervalCount, trialDays int) (*domain.SubscriptionPlan, error) {
	plan, err := domain.NewSubscriptionPlan(merchantID, name, amount, currency, interval, intervalCount)
	if err != nil {
		return nil, err
	}
	plan.Description = description
	plan.TrialDays = trialDays

	if err := s.plans.Create(ctx, plan); err != nil {
		return nil, fmt.Errorf("storing plan: %w", err)
	}
	return plan, nil
}

// GetPlan returns a plan by ID.
func (s *SubscriptionService) GetPlan(ctx context.Context, id uuid.UUID) (*domain.SubscriptionPlan, error) {
	return s.plans.GetByID(ctx, id)
}

// ListPlans returns all plans for a merchant.
func (s *SubscriptionService) ListPlans(ctx context.Context, merchantID uuid.UUID) ([]*domain.SubscriptionPlan, error) {
	return s.plans.ListByMerchant(ctx, merchantID)
}

// ArchivePlan archives a subscription plan.
func (s *SubscriptionService) ArchivePlan(ctx context.Context, id uuid.UUID) error {
	plan, err := s.plans.GetByID(ctx, id)
	if err != nil {
		return err
	}
	plan.Status = domain.PlanArchived
	plan.UpdatedAt = time.Now().UTC()
	return s.plans.Update(ctx, plan)
}

// Subscribe creates a new subscription to a plan.
func (s *SubscriptionService) Subscribe(ctx context.Context, planID uuid.UUID, email, wallet string) (*domain.Subscription, error) {
	plan, err := s.plans.GetByID(ctx, planID)
	if err != nil {
		return nil, err
	}

	sub, err := domain.NewSubscription(planID, plan.MerchantID, email, plan.Amount, plan.IntervalType, plan.IntervalCount, plan.TrialDays)
	if err != nil {
		return nil, err
	}
	sub.SubscriberWallet = wallet

	if err := s.subs.Create(ctx, sub); err != nil {
		return nil, fmt.Errorf("storing subscription: %w", err)
	}
	return sub, nil
}

// GetSubscription returns a subscription by ID.
func (s *SubscriptionService) GetSubscription(ctx context.Context, id uuid.UUID) (*domain.Subscription, error) {
	return s.subs.GetByID(ctx, id)
}

// ListSubscriptions returns all subscriptions for a merchant.
func (s *SubscriptionService) ListSubscriptions(ctx context.Context, merchantID uuid.UUID) ([]*domain.Subscription, error) {
	return s.subs.ListByMerchant(ctx, merchantID)
}

// CancelSubscription cancels a subscription at period end.
func (s *SubscriptionService) CancelSubscription(ctx context.Context, id uuid.UUID, reason string) error {
	sub, err := s.subs.GetByID(ctx, id)
	if err != nil {
		return err
	}
	sub.CancelAtPeriodEnd(reason)
	if err := sub.TransitionTo(domain.SubCancelled); err != nil {
		return err
	}
	return s.subs.Update(ctx, sub)
}
