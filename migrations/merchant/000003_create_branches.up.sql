CREATE TABLE branches (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    merchant_id     UUID NOT NULL REFERENCES merchants(id),
    name            VARCHAR(255) NOT NULL,
    description     TEXT,
    address         TEXT,
    city            VARCHAR(100),
    bank_name       VARCHAR(255),
    bank_branch     VARCHAR(255),
    bank_account_no VARCHAR(50),
    bank_account_name VARCHAR(255),
    is_active       BOOLEAN DEFAULT TRUE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ
);

CREATE INDEX idx_branches_merchant ON branches(merchant_id);
