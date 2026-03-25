CREATE TABLE email_templates (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_type   VARCHAR(100) NOT NULL UNIQUE,
    name         VARCHAR(255) NOT NULL,
    subject      VARCHAR(500) NOT NULL,
    body_html    TEXT NOT NULL,
    variables    JSONB DEFAULT '[]',
    is_active    BOOLEAN NOT NULL DEFAULT TRUE,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Seed default email templates
INSERT INTO email_templates (event_type, name, subject, body_html, variables) VALUES
('kyc.submitted', 'KYC Submitted', 'KYC Application Submitted — Open Pay',
 '<p>Thank you for submitting your KYC application for <strong>{{businessName}}</strong>.</p><p>Our team will review your application and get back to you within 1-3 business days. You have been granted instant access with a limited transaction volume in the meantime.</p><p>If you have any questions, please contact our support team.</p>',
 '["businessName"]'),
('kyc.approved', 'KYC Approved', 'KYC Approved — Open Pay',
 '<p>Congratulations! Your KYC application for <strong>{{businessName}}</strong> has been approved.</p><p>You now have full access to all Open Pay features with no transaction limits. Start accepting payments today!</p>',
 '["businessName"]'),
('kyc.rejected', 'KYC Rejected', 'KYC Application Update — Open Pay',
 '<p>We regret to inform you that your KYC application for <strong>{{businessName}}</strong> was not approved.</p><p><strong>Reason:</strong> {{reason}}</p><p>Please review the feedback and resubmit your application with the required changes. If you have questions, contact our support team.</p>',
 '["businessName", "reason"]'),
('payment.paid', 'Payment Confirmed', 'Payment Confirmed — Open Pay',
 '<p>Payment confirmed!</p><p><strong>Amount:</strong> {{amount}} {{currency}}<br/><strong>Payment No:</strong> {{paymentNo}}</p>',
 '["amount", "currency", "paymentNo"]'),
('payment.expired', 'Payment Expired', 'Payment Expired — Open Pay',
 '<p>Payment <strong>{{paymentNo}}</strong> has expired. No funds were collected.</p>',
 '["paymentNo"]'),
('withdrawal.completed', 'Withdrawal Completed', 'Withdrawal Processed — Open Pay',
 '<p>Your withdrawal of <strong>{{amount}} {{currency}}</strong> has been processed.</p><p><strong>Bank Reference:</strong> {{bankRef}}</p>',
 '["amount", "currency", "bankRef"]');
