ALTER TABLE payments
    ADD COLUMN IF NOT EXISTS customer_first_name VARCHAR(100),
    ADD COLUMN IF NOT EXISTS customer_last_name VARCHAR(100),
    ADD COLUMN IF NOT EXISTS customer_phone VARCHAR(50),
    ADD COLUMN IF NOT EXISTS customer_address TEXT,
    ADD COLUMN IF NOT EXISTS goods JSONB,
    ADD COLUMN IF NOT EXISTS success_url TEXT,
    ADD COLUMN IF NOT EXISTS cancel_url TEXT,
    ADD COLUMN IF NOT EXISTS lkr_amount NUMERIC(20,8),
    ADD COLUMN IF NOT EXISTS lkr_exchange_fee NUMERIC(20,8),
    ADD COLUMN IF NOT EXISTS lkr_platform_fee NUMERIC(20,8),
    ADD COLUMN IF NOT EXISTS lkr_total_fees NUMERIC(20,8),
    ADD COLUMN IF NOT EXISTS lkr_net_amount NUMERIC(20,8);

CREATE INDEX IF NOT EXISTS idx_payments_branch_id ON payments(branch_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_payments_customer_email ON payments(customer_email) WHERE deleted_at IS NULL;
