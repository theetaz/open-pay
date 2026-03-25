ALTER TABLE payment_links
    ADD COLUMN IF NOT EXISTS allow_quantity_buy BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS max_quantity INT NOT NULL DEFAULT 1,
    ADD COLUMN IF NOT EXISTS min_amount NUMERIC(20,8),
    ADD COLUMN IF NOT EXISTS max_amount NUMERIC(20,8),
    ADD COLUMN IF NOT EXISTS success_url TEXT,
    ADD COLUMN IF NOT EXISTS cancel_url TEXT,
    ADD COLUMN IF NOT EXISTS webhook_url TEXT,
    ADD COLUMN IF NOT EXISTS merchant_trade_no VARCHAR(100),
    ADD COLUMN IF NOT EXISTS order_expire_minutes INT;

CREATE INDEX IF NOT EXISTS idx_payment_links_branch_id ON payment_links(branch_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_payment_links_expire_at ON payment_links(expire_at) WHERE status = 'ACTIVE' AND expire_at IS NOT NULL AND deleted_at IS NULL;
