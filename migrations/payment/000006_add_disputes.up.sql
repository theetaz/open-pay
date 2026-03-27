CREATE TABLE disputes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    payment_id UUID NOT NULL,
    merchant_id UUID NOT NULL,
    customer_email VARCHAR(255) NOT NULL,
    reason TEXT NOT NULL,
    evidence_url TEXT,
    status VARCHAR(30) NOT NULL DEFAULT 'OPENED'
        CHECK (status IN ('OPENED', 'MERCHANT_RESPONSE', 'UNDER_REVIEW', 'RESOLVED_CUSTOMER', 'RESOLVED_MERCHANT', 'REJECTED')),
    merchant_response TEXT,
    merchant_responded_at TIMESTAMPTZ,
    admin_notes TEXT,
    resolved_by UUID,
    resolved_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_disputes_payment ON disputes(payment_id);
CREATE INDEX idx_disputes_merchant ON disputes(merchant_id);
CREATE INDEX idx_disputes_status ON disputes(status);
