-- Expand currency constraint to support multiple cryptocurrencies
ALTER TABLE payments DROP CONSTRAINT IF EXISTS payments_currency_check;
ALTER TABLE payments ADD CONSTRAINT payments_currency_check
    CHECK (currency IN ('USDT', 'USDC', 'BTC', 'ETH', 'BNB', 'LKR'));
