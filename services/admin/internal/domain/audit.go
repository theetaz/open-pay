package domain

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidAuditEntry = errors.New("invalid audit log entry")
	ErrAuditNotFound     = errors.New("audit log not found")
)

// Actor types.
const (
	ActorAdmin        = "ADMIN"
	ActorMerchantUser = "MERCHANT_USER"
	ActorSystem       = "SYSTEM"
)

var validActorTypes = map[string]bool{
	ActorAdmin: true, ActorMerchantUser: true, ActorSystem: true,
}

// Change tracks an old and new value for a field.
type Change struct {
	Old string `json:"old"`
	New string `json:"new"`
}

// AuditInput holds data to create an audit entry.
type AuditInput struct {
	ActorID      uuid.UUID
	ActorType    string
	MerchantID   *uuid.UUID
	Action       string
	ResourceType string
	ResourceID   *uuid.UUID
	Changes      map[string]Change
	Metadata     map[string]string
	IPAddress    string
	UserAgent    string
}

// AuditLog represents an immutable audit trail entry.
type AuditLog struct {
	ID           uuid.UUID
	ActorID      uuid.UUID
	ActorType    string
	MerchantID   *uuid.UUID
	Action       string
	ResourceType string
	ResourceID   *uuid.UUID
	Changes      map[string]Change
	Metadata     map[string]string
	IPAddress    string
	UserAgent    string
	CreatedAt    time.Time
}

// NewAuditLog creates a validated audit log entry.
func NewAuditLog(input AuditInput) (*AuditLog, error) {
	if input.Action == "" {
		return nil, fmt.Errorf("%w: action is required", ErrInvalidAuditEntry)
	}
	if !validActorTypes[input.ActorType] {
		return nil, fmt.Errorf("%w: unsupported actor type %s", ErrInvalidAuditEntry, input.ActorType)
	}

	return &AuditLog{
		ID:           uuid.New(),
		ActorID:      input.ActorID,
		ActorType:    input.ActorType,
		MerchantID:   input.MerchantID,
		Action:       input.Action,
		ResourceType: input.ResourceType,
		ResourceID:   input.ResourceID,
		Changes:      input.Changes,
		Metadata:     input.Metadata,
		IPAddress:    input.IPAddress,
		UserAgent:    input.UserAgent,
		CreatedAt:    time.Now().UTC(),
	}, nil
}
