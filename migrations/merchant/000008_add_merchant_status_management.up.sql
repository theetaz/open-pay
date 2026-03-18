-- Add new status values and management columns
ALTER TABLE merchants DROP CONSTRAINT IF EXISTS merchants_status_check;
ALTER TABLE merchants ADD CONSTRAINT merchants_status_check
  CHECK (status IN ('ACTIVE', 'INACTIVE', 'FROZEN', 'TERMINATED'));

ALTER TABLE merchants ADD COLUMN IF NOT EXISTS status_reason TEXT;
ALTER TABLE merchants ADD COLUMN IF NOT EXISTS status_changed_at TIMESTAMPTZ;
ALTER TABLE merchants ADD COLUMN IF NOT EXISTS kyc_review_notes TEXT;
