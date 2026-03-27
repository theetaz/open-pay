package domain

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidDispute           = errors.New("invalid dispute data")
	ErrInvalidDisputeTransition = errors.New("invalid dispute status transition")
	ErrDisputeNotFound          = errors.New("dispute not found")
)

// DisputeStatus represents the state of a dispute.
type DisputeStatus string

const (
	DisputeOpened           DisputeStatus = "OPENED"
	DisputeMerchantResponse DisputeStatus = "MERCHANT_RESPONSE"
	DisputeUnderReview      DisputeStatus = "UNDER_REVIEW"
	DisputeResolvedCustomer DisputeStatus = "RESOLVED_CUSTOMER"
	DisputeResolvedMerchant DisputeStatus = "RESOLVED_MERCHANT"
	DisputeRejected         DisputeStatus = "REJECTED"
)

// validDisputeTransitions defines the allowed state transitions for disputes.
// OPENED → MERCHANT_RESPONSE → UNDER_REVIEW → RESOLVED_CUSTOMER / RESOLVED_MERCHANT / REJECTED
var validDisputeTransitions = map[DisputeStatus][]DisputeStatus{
	DisputeOpened:           {DisputeMerchantResponse},
	DisputeMerchantResponse: {DisputeUnderReview},
	DisputeUnderReview:      {DisputeResolvedCustomer, DisputeResolvedMerchant, DisputeRejected},
}

// Dispute represents a payment dispute raised by a customer.
type Dispute struct {
	ID                  uuid.UUID
	PaymentID           uuid.UUID
	MerchantID          uuid.UUID
	CustomerEmail       string
	Reason              string
	EvidenceURL         string
	Status              DisputeStatus
	MerchantResponse    string
	MerchantRespondedAt *time.Time
	AdminNotes          string
	ResolvedBy          *uuid.UUID
	ResolvedAt          *time.Time
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

// NewDispute creates a validated Dispute from input parameters.
func NewDispute(paymentID, merchantID uuid.UUID, customerEmail, reason string) (*Dispute, error) {
	if paymentID == uuid.Nil {
		return nil, fmt.Errorf("%w: payment ID is required", ErrInvalidDispute)
	}
	if merchantID == uuid.Nil {
		return nil, fmt.Errorf("%w: merchant ID is required", ErrInvalidDispute)
	}
	if customerEmail == "" {
		return nil, fmt.Errorf("%w: customer email is required", ErrInvalidDispute)
	}
	if reason == "" {
		return nil, fmt.Errorf("%w: reason is required", ErrInvalidDispute)
	}

	now := time.Now().UTC()
	return &Dispute{
		ID:            uuid.New(),
		PaymentID:     paymentID,
		MerchantID:    merchantID,
		CustomerEmail: customerEmail,
		Reason:        reason,
		Status:        DisputeOpened,
		CreatedAt:     now,
		UpdatedAt:     now,
	}, nil
}

// TransitionTo moves the dispute to a new status if the transition is valid.
func (d *Dispute) TransitionTo(to DisputeStatus) error {
	allowed, ok := validDisputeTransitions[d.Status]
	if !ok {
		return fmt.Errorf("%w: no transitions from %s", ErrInvalidDisputeTransition, d.Status)
	}

	for _, s := range allowed {
		if s == to {
			d.Status = to
			d.UpdatedAt = time.Now().UTC()
			return nil
		}
	}

	return fmt.Errorf("%w: cannot transition from %s to %s", ErrInvalidDisputeTransition, d.Status, to)
}

// RespondAsMerchant records the merchant's response and transitions to MERCHANT_RESPONSE.
func (d *Dispute) RespondAsMerchant(response string) error {
	if response == "" {
		return fmt.Errorf("%w: merchant response is required", ErrInvalidDispute)
	}
	if err := d.TransitionTo(DisputeMerchantResponse); err != nil {
		return err
	}
	now := time.Now().UTC()
	d.MerchantResponse = response
	d.MerchantRespondedAt = &now
	return nil
}

// Resolve resolves the dispute in favor of either the customer or the merchant.
func (d *Dispute) Resolve(adminID uuid.UUID, notes string, inFavorOfCustomer bool) error {
	target := DisputeResolvedMerchant
	if inFavorOfCustomer {
		target = DisputeResolvedCustomer
	}
	if err := d.TransitionTo(target); err != nil {
		return err
	}
	now := time.Now().UTC()
	d.ResolvedBy = &adminID
	d.ResolvedAt = &now
	d.AdminNotes = notes
	return nil
}

// Reject rejects the dispute (e.g., invalid or fraudulent claim).
func (d *Dispute) Reject(adminID uuid.UUID, notes string) error {
	if err := d.TransitionTo(DisputeRejected); err != nil {
		return err
	}
	now := time.Now().UTC()
	d.ResolvedBy = &adminID
	d.ResolvedAt = &now
	d.AdminNotes = notes
	return nil
}
