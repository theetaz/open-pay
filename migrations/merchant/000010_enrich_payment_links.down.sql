ALTER TABLE payment_links
    DROP COLUMN IF EXISTS allow_quantity_buy,
    DROP COLUMN IF EXISTS max_quantity,
    DROP COLUMN IF EXISTS min_amount,
    DROP COLUMN IF EXISTS max_amount,
    DROP COLUMN IF EXISTS success_url,
    DROP COLUMN IF EXISTS cancel_url,
    DROP COLUMN IF EXISTS webhook_url,
    DROP COLUMN IF EXISTS merchant_trade_no,
    DROP COLUMN IF EXISTS order_expire_minutes;

DROP INDEX IF EXISTS idx_payment_links_branch_id;
DROP INDEX IF EXISTS idx_payment_links_expire_at;
