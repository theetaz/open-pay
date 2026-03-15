CREATE TABLE subscription_plans (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    merchant_id     UUID NOT NULL,
    name            VARCHAR(255) NOT NULL,
    description     TEXT,
    amount          NUMERIC(20,8) NOT NULL CHECK (amount > 0),
    currency        VARCHAR(10) NOT NULL,
    interval_type   VARCHAR(20) NOT NULL CHECK (interval_type IN ('DAILY', 'WEEKLY', 'MONTHLY', 'YEARLY')),
    interval_count  INT NOT NULL DEFAULT 1 CHECK (interval_count > 0),
    trial_days      INT DEFAULT 0,
    max_subscribers INT,
    contract_address VARCHAR(255),
    status          VARCHAR(20) NOT NULL DEFAULT 'ACTIVE' CHECK (status IN ('ACTIVE', 'ARCHIVED')),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ
);

CREATE INDEX idx_plans_merchant ON subscription_plans(merchant_id);
CREATE INDEX idx_plans_status ON subscription_plans(status);

CREATE TABLE subscriptions (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    plan_id             UUID NOT NULL REFERENCES subscription_plans(id),
    merchant_id         UUID NOT NULL,
    subscriber_email    VARCHAR(255) NOT NULL,
    subscriber_wallet   VARCHAR(255),
    status              VARCHAR(20) NOT NULL DEFAULT 'ACTIVE'
                        CHECK (status IN ('TRIAL', 'ACTIVE', 'PAST_DUE', 'PAUSED', 'CANCELLED')),
    current_period_start TIMESTAMPTZ NOT NULL,
    current_period_end   TIMESTAMPTZ NOT NULL,
    next_billing_date    TIMESTAMPTZ,
    trial_end           TIMESTAMPTZ,
    cancel_at_end       BOOLEAN DEFAULT FALSE,
    cancelled_at        TIMESTAMPTZ,
    cancellation_reason TEXT,
    total_paid_usdt     NUMERIC(20,8) DEFAULT 0,
    billing_count       INT DEFAULT 0,
    failed_payment_count INT DEFAULT 0,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_subs_merchant ON subscriptions(merchant_id);
CREATE INDEX idx_subs_plan ON subscriptions(plan_id);
CREATE INDEX idx_subs_status ON subscriptions(status);
CREATE INDEX idx_subs_next_billing ON subscriptions(next_billing_date) WHERE status IN ('ACTIVE', 'TRIAL');
