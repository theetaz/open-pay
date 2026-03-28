-- Tenants table for white-label / multi-tenant support
-- Note: tenant_id on merchants must be added via merchant_db migration, not here
CREATE TABLE tenants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(100) UNIQUE NOT NULL,
    domain VARCHAR(255),
    logo_url TEXT,
    primary_color VARCHAR(7) DEFAULT '#6366F1',
    platform_fee_pct NUMERIC(5,2) DEFAULT 1.5,
    exchange_fee_pct NUMERIC(5,2) DEFAULT 0.5,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_tenants_slug ON tenants(slug);
CREATE INDEX idx_tenants_domain ON tenants(domain);
