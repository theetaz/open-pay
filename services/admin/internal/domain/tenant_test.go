package domain_test

import (
	"testing"

	"github.com/openlankapay/openlankapay/services/admin/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTenant(t *testing.T) {
	t.Run("valid tenant with defaults", func(t *testing.T) {
		tenant, err := domain.NewTenant(domain.NewTenantInput{
			Name:           "Acme Corp",
			Slug:           "acme-corp",
			PrimaryColor:   "#FF5733",
			PlatformFeePct: 2.0,
			ExchangeFeePct: 0.5,
		})
		require.NoError(t, err)
		assert.NotEmpty(t, tenant.ID)
		assert.Equal(t, "Acme Corp", tenant.Name)
		assert.Equal(t, "acme-corp", tenant.Slug)
		assert.Equal(t, "#FF5733", tenant.PrimaryColor)
		assert.Equal(t, 2.0, tenant.PlatformFeePct)
		assert.Equal(t, 0.5, tenant.ExchangeFeePct)
		assert.True(t, tenant.IsActive)
		assert.Nil(t, tenant.Domain)
		assert.Nil(t, tenant.LogoURL)
		assert.False(t, tenant.CreatedAt.IsZero())
		assert.False(t, tenant.UpdatedAt.IsZero())
	})

	t.Run("valid tenant with optional fields", func(t *testing.T) {
		domainStr := "acme.example.com"
		logoURL := "https://cdn.example.com/acme-logo.png"
		tenant, err := domain.NewTenant(domain.NewTenantInput{
			Name:           "Acme Corp",
			Slug:           "acme-corp",
			Domain:         &domainStr,
			LogoURL:        &logoURL,
			PrimaryColor:   "#6366F1",
			PlatformFeePct: 1.5,
			ExchangeFeePct: 0.5,
		})
		require.NoError(t, err)
		require.NotNil(t, tenant.Domain)
		assert.Equal(t, "acme.example.com", *tenant.Domain)
		require.NotNil(t, tenant.LogoURL)
		assert.Equal(t, "https://cdn.example.com/acme-logo.png", *tenant.LogoURL)
	})

	t.Run("default primary color when empty", func(t *testing.T) {
		tenant, err := domain.NewTenant(domain.NewTenantInput{
			Name: "Defaults",
			Slug: "defaults",
		})
		require.NoError(t, err)
		assert.Equal(t, "#6366F1", tenant.PrimaryColor)
	})

	t.Run("empty name is invalid", func(t *testing.T) {
		_, err := domain.NewTenant(domain.NewTenantInput{
			Name: "",
			Slug: "test",
		})
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrInvalidTenant)
	})

	t.Run("empty slug is invalid", func(t *testing.T) {
		_, err := domain.NewTenant(domain.NewTenantInput{
			Name: "Test",
			Slug: "",
		})
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrInvalidTenant)
	})

	t.Run("slug with uppercase is invalid", func(t *testing.T) {
		_, err := domain.NewTenant(domain.NewTenantInput{
			Name: "Test",
			Slug: "Acme-Corp",
		})
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrInvalidTenant)
	})

	t.Run("slug with spaces is invalid", func(t *testing.T) {
		_, err := domain.NewTenant(domain.NewTenantInput{
			Name: "Test",
			Slug: "acme corp",
		})
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrInvalidTenant)
	})

	t.Run("slug with trailing hyphen is invalid", func(t *testing.T) {
		_, err := domain.NewTenant(domain.NewTenantInput{
			Name: "Test",
			Slug: "acme-",
		})
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrInvalidTenant)
	})

	t.Run("invalid hex color", func(t *testing.T) {
		_, err := domain.NewTenant(domain.NewTenantInput{
			Name:         "Test",
			Slug:         "test",
			PrimaryColor: "red",
		})
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrInvalidTenant)
	})

	t.Run("negative platform fee is invalid", func(t *testing.T) {
		_, err := domain.NewTenant(domain.NewTenantInput{
			Name:           "Test",
			Slug:           "test",
			PlatformFeePct: -1.0,
		})
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrInvalidTenant)
	})

	t.Run("negative exchange fee is invalid", func(t *testing.T) {
		_, err := domain.NewTenant(domain.NewTenantInput{
			Name:           "Test",
			Slug:           "test",
			ExchangeFeePct: -0.5,
		})
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrInvalidTenant)
	})
}

func TestTenantSlugValidation(t *testing.T) {
	tests := []struct {
		slug  string
		valid bool
	}{
		{"acme", true},
		{"acme-corp", true},
		{"my-tenant-123", true},
		{"a", true},
		{"123", true},
		{"ACME", false},
		{"acme_corp", false},
		{"acme corp", false},
		{"-acme", false},
		{"acme-", false},
		{"acme--corp", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.slug, func(t *testing.T) {
			_, err := domain.NewTenant(domain.NewTenantInput{
				Name: "Test",
				Slug: tt.slug,
			})
			if tt.valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}
