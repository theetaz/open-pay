DROP TABLE IF EXISTS watcher_cursor;
DROP INDEX IF EXISTS idx_payments_onchain_pending;
ALTER TABLE payments DROP COLUMN IF EXISTS deposit_index;
DROP SEQUENCE IF EXISTS payment_deposit_index_seq;

ALTER TABLE payments DROP CONSTRAINT IF EXISTS payments_provider_check;
ALTER TABLE payments ADD CONSTRAINT payments_provider_check
    CHECK (provider IN ('BYBIT', 'BINANCE', 'KUCOIN', 'TEST'));
