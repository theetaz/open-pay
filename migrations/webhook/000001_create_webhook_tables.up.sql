CREATE TABLE webhook_configs (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    merchant_id         UUID UNIQUE NOT NULL,
    url                 TEXT NOT NULL,
    signing_public_key  TEXT NOT NULL,
    signing_private_key TEXT NOT NULL,
    events              JSONB DEFAULT '["payment.*"]',
    is_active           BOOLEAN DEFAULT TRUE,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE webhook_deliveries (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    webhook_config_id   UUID NOT NULL REFERENCES webhook_configs(id),
    merchant_id         UUID NOT NULL,
    event_type          VARCHAR(50) NOT NULL,
    payload             JSONB NOT NULL,
    attempt_count       INT DEFAULT 0,
    max_attempts        INT DEFAULT 5,
    status              VARCHAR(20) NOT NULL DEFAULT 'PENDING'
                        CHECK (status IN ('PENDING', 'DELIVERING', 'DELIVERED', 'FAILED', 'EXHAUSTED')),
    last_response_code  INT,
    last_response_body  TEXT,
    last_error          TEXT,
    next_attempt_at     TIMESTAMPTZ,
    delivered_at        TIMESTAMPTZ,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_deliveries_status ON webhook_deliveries(status);
CREATE INDEX idx_deliveries_next_attempt ON webhook_deliveries(next_attempt_at) WHERE status = 'PENDING';
CREATE INDEX idx_deliveries_merchant ON webhook_deliveries(merchant_id);
