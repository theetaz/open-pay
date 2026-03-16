DROP INDEX IF EXISTS idx_payment_links_slug;
CREATE UNIQUE INDEX idx_payment_links_merchant_slug ON payment_links (merchant_id, slug) WHERE deleted_at IS NULL;
