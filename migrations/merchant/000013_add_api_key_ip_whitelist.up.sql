-- Add IP whitelist to API keys (null = allow all IPs)
ALTER TABLE api_keys ADD COLUMN IF NOT EXISTS allowed_ips TEXT[];
