CREATE INDEX IF NOT EXISTS idx_deliveries_retryable
    ON webhook_deliveries(next_attempt_at)
    WHERE status = 'PENDING' AND next_attempt_at IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_deliveries_merchant
    ON webhook_deliveries(merchant_id, created_at DESC);
