CREATE TABLE platform_settings (
    key         VARCHAR(100) PRIMARY KEY,
    value       TEXT NOT NULL,
    description TEXT,
    category    VARCHAR(50) NOT NULL DEFAULT 'general',
    updated_by  UUID REFERENCES admin_users(id),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Seed default settings
INSERT INTO platform_settings (key, value, description, category) VALUES
('platform_name', 'Open Pay', 'Platform display name', 'general'),
('platform_logo_url', '', 'Platform logo URL', 'general'),
('support_email', 'support@openpay.lk', 'Support email address', 'general'),
('platform_fee_pct', '1.5', 'Platform fee percentage', 'fees'),
('exchange_fee_pct', '0.5', 'Exchange fee percentage', 'fees'),
('min_withdrawal_usdt', '10', 'Minimum withdrawal amount in USDT', 'fees'),
('settlement_period_days', '3', 'Settlement period in business days', 'fees'),
('smtp_host', 'localhost', 'SMTP server host', 'email'),
('smtp_port', '1025', 'SMTP server port', 'email'),
('smtp_sender_name', 'Open Pay', 'Email sender name', 'email'),
('smtp_sender_email', 'noreply@openlankapay.dev', 'Email sender address', 'email');
