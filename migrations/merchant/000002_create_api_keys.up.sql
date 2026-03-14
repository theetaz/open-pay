CREATE TABLE api_keys (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    merchant_id     UUID NOT NULL REFERENCES merchants(id),
    key_id          VARCHAR(100) UNIQUE NOT NULL,
    secret_hash     VARCHAR(255) NOT NULL,
    name            VARCHAR(255),
    environment     VARCHAR(10) NOT NULL CHECK (environment IN ('live', 'test')),
    is_active       BOOLEAN DEFAULT TRUE,
    revoked_at      TIMESTAMPTZ,
    revoked_reason  TEXT,
    last_used_at    TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_api_keys_merchant ON api_keys(merchant_id);
CREATE INDEX idx_api_keys_key_id ON api_keys(key_id);
CREATE INDEX idx_api_keys_active ON api_keys(merchant_id, is_active) WHERE is_active = TRUE;
