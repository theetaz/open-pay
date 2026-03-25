ALTER TABLE merchants DROP CONSTRAINT IF EXISTS merchants_status_check;
ALTER TABLE merchants ADD CONSTRAINT merchants_status_check
  CHECK (status IN ('ACTIVE', 'INACTIVE'));

ALTER TABLE merchants DROP COLUMN IF EXISTS status_reason;
ALTER TABLE merchants DROP COLUMN IF EXISTS status_changed_at;
ALTER TABLE merchants DROP COLUMN IF EXISTS kyc_review_notes;
