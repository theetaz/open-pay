-- Make slug globally unique (not just per-merchant) for clean public URLs
DROP INDEX IF EXISTS idx_payment_links_merchant_slug;
CREATE UNIQUE INDEX idx_payment_links_slug ON payment_links (slug) WHERE deleted_at IS NULL;
