-- Scenario codes define provider-specific pre-authorization limits and types.
CREATE TABLE scenario_codes (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    scenario_id     VARCHAR(64) NOT NULL,
    scenario_name   VARCHAR(255) NOT NULL,
    payment_provider VARCHAR(20) NOT NULL CHECK (payment_provider IN ('BINANCE_PAY', 'BYBIT_PAY', 'KUCOIN_PAY')),
    max_limit       NUMERIC(20,8) NOT NULL CHECK (max_limit > 0),
    is_active       BOOLEAN NOT NULL DEFAULT TRUE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_scenario_provider_id ON scenario_codes(payment_provider, scenario_id);
CREATE INDEX idx_scenario_active ON scenario_codes(is_active);

-- Direct debit contracts represent pre-authorized agreements between merchant and customer.
CREATE TABLE direct_debit_contracts (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    merchant_id             UUID NOT NULL,
    branch_id               UUID,
    merchant_contract_code  VARCHAR(64) NOT NULL,
    service_name            VARCHAR(64) NOT NULL,
    scenario_id             UUID NOT NULL REFERENCES scenario_codes(id),
    payment_provider        VARCHAR(20) NOT NULL,
    currency                VARCHAR(10) NOT NULL DEFAULT 'USDT',
    single_upper_limit      NUMERIC(20,8) NOT NULL CHECK (single_upper_limit >= 0.01),
    status                  VARCHAR(20) NOT NULL DEFAULT 'INITIATED'
                            CHECK (status IN ('INITIATED', 'SIGNED', 'TERMINATED', 'EXPIRED')),
    -- Provider-specific identifiers
    pre_contract_id         VARCHAR(255),
    contract_id             VARCHAR(255),
    biz_id                  VARCHAR(255),
    open_user_id            VARCHAR(255),
    merchant_account_no     VARCHAR(255),
    -- QR and deep link for user signing
    qr_content              TEXT,
    deep_link               TEXT,
    -- URLs
    webhook_url             VARCHAR(512),
    return_url              VARCHAR(512) NOT NULL,
    cancel_url              VARCHAR(512) NOT NULL,
    -- Contract lifecycle
    periodic                BOOLEAN DEFAULT FALSE,
    contract_end_time       TIMESTAMPTZ,
    request_expire_time     TIMESTAMPTZ,
    termination_way         INT,
    termination_time        TIMESTAMPTZ,
    termination_notes       VARCHAR(256),
    -- Aggregated stats
    payment_count           INT NOT NULL DEFAULT 0,
    total_amount_charged    NUMERIC(20,8) NOT NULL DEFAULT 0,
    last_payment_at         TIMESTAMPTZ,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_contract_merchant_code ON direct_debit_contracts(merchant_id, merchant_contract_code);
CREATE INDEX idx_contract_merchant ON direct_debit_contracts(merchant_id);
CREATE INDEX idx_contract_status ON direct_debit_contracts(status);
CREATE INDEX idx_contract_provider ON direct_debit_contracts(payment_provider);

-- Direct debit payments are charges executed against signed contracts.
CREATE TABLE direct_debit_payments (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    contract_id             UUID NOT NULL REFERENCES direct_debit_contracts(id),
    merchant_id             UUID NOT NULL,
    pay_id                  VARCHAR(255),
    payment_no              VARCHAR(64) NOT NULL,
    amount                  NUMERIC(20,8) NOT NULL CHECK (amount >= 0.01),
    currency                VARCHAR(10) NOT NULL DEFAULT 'USDT',
    status                  VARCHAR(20) NOT NULL DEFAULT 'INITIATED'
                            CHECK (status IN ('INITIATED', 'PAID', 'FAILED', 'REFUNDED')),
    product_name            VARCHAR(256) NOT NULL,
    product_detail          VARCHAR(256),
    payment_provider        VARCHAR(20) NOT NULL,
    webhook_url             VARCHAR(512),
    -- Fee breakdown
    gross_amount_usdt       NUMERIC(20,8),
    exchange_fee_pct        NUMERIC(5,2),
    exchange_fee_usdt       NUMERIC(20,8),
    platform_fee_pct        NUMERIC(5,2),
    platform_fee_usdt       NUMERIC(20,8),
    total_fees_usdt         NUMERIC(20,8),
    net_amount_usdt         NUMERIC(20,8),
    -- Customer billing info
    customer_first_name     VARCHAR(100),
    customer_last_name      VARCHAR(100),
    customer_email          VARCHAR(255),
    customer_phone          VARCHAR(50),
    customer_address        TEXT,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_dd_payment_no ON direct_debit_payments(payment_no);
CREATE INDEX idx_dd_payment_contract ON direct_debit_payments(contract_id);
CREATE INDEX idx_dd_payment_merchant ON direct_debit_payments(merchant_id);
CREATE INDEX idx_dd_payment_status ON direct_debit_payments(status);
