CREATE TABLE payment_links (
    id                  UUID PRIMARY KEY,
    merchant_id         UUID NOT NULL,
    branch_id           UUID,
    name                VARCHAR(255) NOT NULL,
    slug                VARCHAR(255) NOT NULL,
    description         TEXT,
    currency            VARCHAR(10) NOT NULL DEFAULT 'LKR',
    amount              NUMERIC(20,8) NOT NULL,
    allow_custom_amount BOOLEAN NOT NULL DEFAULT FALSE,
    is_reusable         BOOLEAN NOT NULL DEFAULT FALSE,
    show_on_qr_page     BOOLEAN NOT NULL DEFAULT FALSE,
    usage_count         INT NOT NULL DEFAULT 0,
    status              VARCHAR(20) NOT NULL DEFAULT 'ACTIVE',
    expire_at           TIMESTAMPTZ,
    created_at          TIMESTAMPTZ NOT NULL,
    updated_at          TIMESTAMPTZ NOT NULL,
    deleted_at          TIMESTAMPTZ
);

CREATE UNIQUE INDEX idx_payment_links_merchant_slug ON payment_links (merchant_id, slug) WHERE deleted_at IS NULL;
CREATE INDEX idx_payment_links_merchant_id ON payment_links (merchant_id);
CREATE INDEX idx_payment_links_status ON payment_links (status);
CREATE INDEX idx_payment_links_not_deleted ON payment_links (id) WHERE deleted_at IS NULL;
