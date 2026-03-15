package domain_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/openlankapay/openlankapay/services/admin/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAuditLog(t *testing.T) {
	t.Run("valid audit entry", func(t *testing.T) {
		resourceID := uuid.New()
		entry, err := domain.NewAuditLog(domain.AuditInput{
			ActorID:      uuid.New(),
			ActorType:    domain.ActorAdmin,
			Action:       "merchant.approved",
			ResourceType: "merchant",
			ResourceID:   &resourceID,
			IPAddress:    "192.168.1.1",
		})
		require.NoError(t, err)
		assert.Equal(t, "merchant.approved", entry.Action)
		assert.Equal(t, domain.ActorAdmin, entry.ActorType)
		assert.NotEmpty(t, entry.ID)
	})

	t.Run("empty action invalid", func(t *testing.T) {
		_, err := domain.NewAuditLog(domain.AuditInput{
			ActorID:   uuid.New(),
			ActorType: domain.ActorAdmin,
			Action:    "",
		})
		require.Error(t, err)
	})

	t.Run("with changes tracking", func(t *testing.T) {
		entry, _ := domain.NewAuditLog(domain.AuditInput{
			ActorID:      uuid.New(),
			ActorType:    domain.ActorAdmin,
			Action:       "merchant.updated",
			ResourceType: "merchant",
			Changes: map[string]domain.Change{
				"kyc_status": {Old: "PENDING", New: "APPROVED"},
			},
		})
		assert.NotNil(t, entry.Changes)
		assert.Equal(t, "PENDING", entry.Changes["kyc_status"].Old)
		assert.Equal(t, "APPROVED", entry.Changes["kyc_status"].New)
	})
}

func TestAuditActorTypes(t *testing.T) {
	tests := []struct {
		actorType string
		valid     bool
	}{
		{domain.ActorAdmin, true},
		{domain.ActorMerchantUser, true},
		{domain.ActorSystem, true},
		{"UNKNOWN", false},
	}

	for _, tt := range tests {
		t.Run(tt.actorType, func(t *testing.T) {
			_, err := domain.NewAuditLog(domain.AuditInput{
				ActorID:   uuid.New(),
				ActorType: tt.actorType,
				Action:    "test.action",
			})
			if tt.valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}
