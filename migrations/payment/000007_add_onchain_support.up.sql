-- On-chain "scan-and-send" payment support.

-- 1. Allow the ONCHAIN provider (the base CHECK only listed CEX providers).
ALTER TABLE payments DROP CONSTRAINT IF EXISTS payments_provider_check;
ALTER TABLE payments ADD CONSTRAINT payments_provider_check
    CHECK (provider IN ('BYBIT', 'BINANCE', 'KUCOIN', 'TEST', 'ONCHAIN'));

-- 2. Per-deposit HD wallet index (unique, monotonic) so each on-chain payment
--    gets its own deterministic deposit address.
CREATE SEQUENCE IF NOT EXISTS payment_deposit_index_seq;
ALTER TABLE payments ADD COLUMN IF NOT EXISTS deposit_index BIGINT;

-- 3. Fast lookup of open on-chain deposits by address (used by the watcher).
CREATE INDEX IF NOT EXISTS idx_payments_onchain_pending
    ON payments (wallet_address)
    WHERE provider = 'ONCHAIN' AND status = 'INITIATED';

-- 4. Watcher block cursor (single row, survives restarts).
CREATE TABLE IF NOT EXISTS watcher_cursor (
    id                 TEXT PRIMARY KEY,
    last_scanned_block BIGINT NOT NULL,
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
