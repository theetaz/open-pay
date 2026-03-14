CREATE TABLE merchants (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    business_name       VARCHAR(255) NOT NULL,
    business_type       VARCHAR(50),
    registration_no     VARCHAR(100),
    website             VARCHAR(500),
    contact_email       VARCHAR(255) NOT NULL UNIQUE,
    contact_phone       VARCHAR(50),
    contact_name        VARCHAR(255),
    address_line1       VARCHAR(500),
    address_line2       VARCHAR(500),
    city                VARCHAR(100),
    district            VARCHAR(100),
    postal_code         VARCHAR(20),
    country             VARCHAR(5) NOT NULL DEFAULT 'LK',
    kyc_status          VARCHAR(20) NOT NULL DEFAULT 'PENDING'
                        CHECK (kyc_status IN ('PENDING', 'UNDER_REVIEW', 'APPROVED', 'REJECTED', 'INSTANT_ACCESS')),
    kyc_submitted_at    TIMESTAMPTZ,
    kyc_reviewed_at     TIMESTAMPTZ,
    kyc_rejection_reason TEXT,
    transaction_limit_usdt NUMERIC(20,8),
    daily_limit_usdt       NUMERIC(20,8),
    monthly_limit_usdt     NUMERIC(20,8),
    instant_access_remaining NUMERIC(20,8) DEFAULT 5000,
    bank_name           VARCHAR(255),
    bank_branch         VARCHAR(255),
    bank_account_no     VARCHAR(50),
    bank_account_name   VARCHAR(255),
    default_currency    VARCHAR(10) DEFAULT 'LKR',
    default_provider    VARCHAR(20),
    status              VARCHAR(20) NOT NULL DEFAULT 'ACTIVE'
                        CHECK (status IN ('ACTIVE', 'INACTIVE')),
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at          TIMESTAMPTZ
);

CREATE INDEX idx_merchants_email ON merchants(contact_email);
CREATE INDEX idx_merchants_kyc_status ON merchants(kyc_status);
CREATE INDEX idx_merchants_status ON merchants(status);
CREATE INDEX idx_merchants_not_deleted ON merchants(id) WHERE deleted_at IS NULL;
