ALTER TABLE payments
    DROP COLUMN IF EXISTS customer_first_name,
    DROP COLUMN IF EXISTS customer_last_name,
    DROP COLUMN IF EXISTS customer_phone,
    DROP COLUMN IF EXISTS customer_address,
    DROP COLUMN IF EXISTS goods,
    DROP COLUMN IF EXISTS success_url,
    DROP COLUMN IF EXISTS cancel_url,
    DROP COLUMN IF EXISTS lkr_amount,
    DROP COLUMN IF EXISTS lkr_exchange_fee,
    DROP COLUMN IF EXISTS lkr_platform_fee,
    DROP COLUMN IF EXISTS lkr_total_fees,
    DROP COLUMN IF EXISTS lkr_net_amount;

DROP INDEX IF EXISTS idx_payments_branch_id;
DROP INDEX IF EXISTS idx_payments_customer_email;
