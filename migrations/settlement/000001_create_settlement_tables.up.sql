CREATE TABLE merchant_balances (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    merchant_id         UUID UNIQUE NOT NULL,
    available_usdt      NUMERIC(20,8) NOT NULL DEFAULT 0 CHECK (available_usdt >= 0),
    pending_usdt        NUMERIC(20,8) NOT NULL DEFAULT 0,
    total_earned_usdt   NUMERIC(20,8) NOT NULL DEFAULT 0,
    total_withdrawn_usdt NUMERIC(20,8) NOT NULL DEFAULT 0,
    total_fees_usdt     NUMERIC(20,8) NOT NULL DEFAULT 0,
    total_earned_lkr    NUMERIC(20,4) NOT NULL DEFAULT 0,
    total_withdrawn_lkr NUMERIC(20,4) NOT NULL DEFAULT 0,
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_balances_merchant ON merchant_balances(merchant_id);

CREATE TABLE withdrawals (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    merchant_id         UUID NOT NULL,
    amount_usdt         NUMERIC(20,8) NOT NULL CHECK (amount_usdt > 0),
    exchange_rate       NUMERIC(20,8) NOT NULL,
    amount_lkr          NUMERIC(20,4) NOT NULL,
    bank_name           VARCHAR(255) NOT NULL,
    bank_account_no     VARCHAR(50) NOT NULL,
    bank_account_name   VARCHAR(255) NOT NULL,
    status              VARCHAR(20) NOT NULL DEFAULT 'REQUESTED'
                        CHECK (status IN ('REQUESTED', 'APPROVED', 'PROCESSING', 'COMPLETED', 'REJECTED')),
    approved_by         UUID,
    approved_at         TIMESTAMPTZ,
    rejected_reason     TEXT,
    bank_reference      VARCHAR(100),
    completed_at        TIMESTAMPTZ,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_withdrawals_merchant ON withdrawals(merchant_id);
CREATE INDEX idx_withdrawals_status ON withdrawals(status);
