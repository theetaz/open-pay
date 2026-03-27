-- Add fraud risk assessment fields to payments
ALTER TABLE payments ADD COLUMN IF NOT EXISTS risk_score INT NOT NULL DEFAULT 0;
ALTER TABLE payments ADD COLUMN IF NOT EXISTS risk_flags TEXT[];
