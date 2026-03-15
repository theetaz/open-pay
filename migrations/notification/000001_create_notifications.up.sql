CREATE TABLE notifications (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    merchant_id     UUID NOT NULL,
    channel         VARCHAR(20) NOT NULL CHECK (channel IN ('EMAIL', 'SMS', 'PUSH')),
    recipient       VARCHAR(255) NOT NULL,
    subject         VARCHAR(500),
    body            TEXT NOT NULL,
    event_type      VARCHAR(50) NOT NULL,
    reference_id    UUID,
    status          VARCHAR(20) NOT NULL DEFAULT 'PENDING'
                    CHECK (status IN ('PENDING', 'SENT', 'FAILED')),
    failure_reason  TEXT,
    sent_at         TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_notifications_merchant ON notifications(merchant_id);
CREATE INDEX idx_notifications_status ON notifications(status);
CREATE INDEX idx_notifications_event_type ON notifications(event_type);
