CREATE TABLE legal_documents (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    type         VARCHAR(50) NOT NULL,
    version      INTEGER NOT NULL,
    title        VARCHAR(500) NOT NULL,
    content      TEXT NOT NULL,
    is_active    BOOLEAN NOT NULL DEFAULT FALSE,
    created_by   UUID,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(type, version)
);

CREATE INDEX idx_legal_docs_active ON legal_documents(type, is_active) WHERE is_active = TRUE;

-- Seed default terms and conditions (version 1)
INSERT INTO legal_documents (type, version, title, content, is_active) VALUES (
    'terms_and_conditions',
    1,
    'Open Pay Payment Gateway — Terms and Conditions',
    'OPEN PAY PAYMENT GATEWAY — TERMS AND CONDITIONS

Last Updated: March 2026

1. INTRODUCTION
These Terms and Conditions ("Agreement") govern your use of the Open Pay Payment Gateway ("Service") operated by Open Lanka Payment (Pvt) Ltd ("Company"), a registered entity under the laws of the Democratic Socialist Republic of Sri Lanka. By accessing or using our Service, you agree to be bound by these terms.

2. DEFINITIONS
"Merchant" refers to any individual or business entity that registers to use the Open Pay Payment Gateway to accept cryptocurrency payments.
"Transaction" refers to any payment processed through the Service.
"Settlement" refers to the conversion and transfer of funds to the Merchant''s designated bank account.

3. ELIGIBILITY
To use the Service, you must:
(a) Be a registered business entity or sole proprietorship in Sri Lanka;
(b) Maintain a valid bank account with a licensed commercial bank in Sri Lanka;
(c) Comply with all applicable laws, including the Payment Devices Fraud Act No. 30 of 2006 and relevant Central Bank of Sri Lanka (CBSL) regulations;
(d) Complete the Know Your Customer (KYC) verification process.

4. PAYMENT PROCESSING
The Company facilitates cryptocurrency payment processing, converting digital assets to fiat currency (LKR or USD) as per the Merchant''s preference. Settlement periods are typically 1–3 business days following transaction confirmation on the respective blockchain network.

5. FEES AND CHARGES
Transaction fees are calculated as a percentage of each processed payment and are deducted at the time of settlement. The current fee schedule is available on the Open Pay dashboard. The Company reserves the right to modify fees with 30 days'' prior written notice.

6. COMPLIANCE AND ANTI-MONEY LAUNDERING
Merchants must comply with the Financial Transactions Reporting Act No. 6 of 2006 and the Prevention of Money Laundering Act No. 5 of 2006. The Company reserves the right to suspend or terminate accounts suspected of involvement in money laundering, terrorist financing, or other illegal activities.

7. DATA PROTECTION
The Company processes personal data in accordance with the Personal Data Protection Act No. 9 of 2022 of Sri Lanka. Merchant data is encrypted and stored securely in compliance with industry standards.

8. LIMITATION OF LIABILITY
The Company shall not be liable for any indirect, incidental, or consequential damages arising from the use of the Service, including but not limited to losses due to cryptocurrency price volatility, network congestion, or blockchain-related delays.

9. TERMINATION
Either party may terminate this Agreement with 30 days'' written notice. The Company may immediately suspend or terminate access if the Merchant breaches any provision of this Agreement or applicable law.

10. GOVERNING LAW
This Agreement shall be governed by and construed in accordance with the laws of the Democratic Socialist Republic of Sri Lanka. Any disputes shall be subject to the exclusive jurisdiction of the courts of Sri Lanka.

11. AMENDMENTS
The Company reserves the right to amend these Terms and Conditions at any time. Continued use of the Service following any amendment constitutes acceptance of the modified terms.

For questions regarding these Terms and Conditions, contact: legal@openpay.lk',
    TRUE
);
