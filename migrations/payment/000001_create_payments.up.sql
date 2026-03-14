CREATE TABLE payments (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    merchant_id         UUID NOT NULL,
    branch_id           UUID,
    payment_no          VARCHAR(50) UNIQUE NOT NULL,
    merchant_trade_no   VARCHAR(100),
    amount              NUMERIC(20,8) NOT NULL CHECK (amount > 0),
    currency            VARCHAR(10) NOT NULL CHECK (currency IN ('USDT', 'LKR')),
    amount_usdt         NUMERIC(20,8) NOT NULL,
    exchange_rate_snapshot NUMERIC(20,8),
    exchange_fee_pct    NUMERIC(5,4),
    exchange_fee_usdt   NUMERIC(20,8),
    platform_fee_pct    NUMERIC(5,4),
    platform_fee_usdt   NUMERIC(20,8),
    total_fees_usdt     NUMERIC(20,8),
    net_amount_usdt     NUMERIC(20,8),
    provider            VARCHAR(20) NOT NULL CHECK (provider IN ('BYBIT', 'BINANCE', 'KUCOIN', 'TEST')),
    provider_pay_id     VARCHAR(255),
    qr_content          TEXT,
    checkout_link       TEXT,
    deep_link           TEXT,
    status              VARCHAR(20) NOT NULL DEFAULT 'INITIATED'
                        CHECK (status IN ('INITIATED', 'USER_REVIEW', 'PAID', 'EXPIRED', 'FAILED')),
    customer_email      VARCHAR(255),
    webhook_url         TEXT,
    tx_hash             VARCHAR(255),
    block_number        BIGINT,
    wallet_address      VARCHAR(255),
    expire_time         TIMESTAMPTZ NOT NULL,
    paid_at             TIMESTAMPTZ,
    failed_at           TIMESTAMPTZ,
    idempotency_key     VARCHAR(255) UNIQUE,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at          TIMESTAMPTZ
);

CREATE INDEX idx_payments_merchant_id ON payments(merchant_id);
CREATE INDEX idx_payments_status ON payments(status);
CREATE INDEX idx_payments_created_at ON payments(created_at);
CREATE INDEX idx_payments_merchant_trade_no ON payments(merchant_trade_no);
CREATE INDEX idx_payments_provider ON payments(provider);
CREATE INDEX idx_payments_expire_time ON payments(expire_time) WHERE status = 'INITIATED';
CREATE INDEX idx_payments_not_deleted ON payments(id) WHERE deleted_at IS NULL;
