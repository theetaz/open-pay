CREATE TABLE exchange_rates (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    base_currency   VARCHAR(10) NOT NULL,
    quote_currency  VARCHAR(10) NOT NULL,
    rate            NUMERIC(20,8) NOT NULL CHECK (rate > 0),
    source          VARCHAR(50) NOT NULL,
    is_active       BOOLEAN DEFAULT TRUE,
    fetched_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_rates_pair ON exchange_rates(base_currency, quote_currency);
CREATE INDEX idx_rates_active ON exchange_rates(is_active, base_currency, quote_currency);
CREATE INDEX idx_rates_fetched ON exchange_rates(fetched_at);
