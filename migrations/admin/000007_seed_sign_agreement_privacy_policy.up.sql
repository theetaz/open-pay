-- Seed default sign agreement (version 1)
INSERT INTO legal_documents (type, version, title, content, is_active) VALUES (
    'sign_agreement',
    1,
    'Open Pay — Merchant Sign Agreement',
    'OPEN PAY — MERCHANT SIGN AGREEMENT

Effective Date: Upon execution by the Merchant

This Merchant Sign Agreement ("Agreement") is entered into between Open Lanka Payment (Pvt) Ltd ("Open Pay", "Company", "we", "us") and the undersigned merchant ("Merchant", "you").

1. ENGAGEMENT
By signing this Agreement, you agree to use the Open Pay Payment Gateway to accept cryptocurrency payments from your customers. Open Pay will provide payment processing, currency conversion, and settlement services as described herein.

2. MERCHANT OBLIGATIONS
The Merchant agrees to:
(a) Provide accurate and complete business information during registration and KYC verification;
(b) Maintain compliance with all applicable laws and regulations of Sri Lanka;
(c) Not use the Service for any illegal, fraudulent, or prohibited activities;
(d) Promptly notify Open Pay of any changes to business information, ownership, or banking details;
(e) Display the Open Pay payment option clearly on their platform or point of sale;
(f) Maintain adequate records of all transactions processed through the Service.

3. PAYMENT PROCESSING
(a) Open Pay will process cryptocurrency payments on behalf of the Merchant;
(b) Settlements will be made in the Merchant''s preferred currency (LKR or USD) to their designated bank account;
(c) Standard settlement period is 1–3 business days following blockchain confirmation;
(d) Open Pay reserves the right to hold settlements if suspicious activity is detected.

4. FEES
(a) Transaction fees are calculated as a percentage of each processed payment;
(b) Current fee rates are published on the Open Pay dashboard and may be updated with 30 days'' notice;
(c) Fees are deducted at the time of settlement;
(d) Additional fees may apply for chargebacks, refunds, or currency conversion.

5. REPRESENTATIONS AND WARRANTIES
The Merchant represents and warrants that:
(a) It is a legally registered business entity in Sri Lanka;
(b) It has the authority to enter into this Agreement;
(c) All information provided during registration and KYC is true, accurate, and complete;
(d) It will comply with the Prevention of Money Laundering Act No. 5 of 2006;
(e) It will comply with the Financial Transactions Reporting Act No. 6 of 2006.

6. INTELLECTUAL PROPERTY
(a) Open Pay retains all rights to its trademarks, logos, and proprietary technology;
(b) The Merchant is granted a limited, non-exclusive license to use Open Pay branding solely for payment acceptance purposes;
(c) Neither party may use the other''s intellectual property without prior written consent.

7. CONFIDENTIALITY
Both parties agree to maintain the confidentiality of proprietary information, customer data, and transaction details. This obligation survives termination of this Agreement.

8. INDEMNIFICATION
The Merchant agrees to indemnify and hold harmless Open Pay from any claims, damages, or losses arising from:
(a) The Merchant''s breach of this Agreement;
(b) The Merchant''s violation of any applicable law;
(c) Any dispute between the Merchant and their customers.

9. LIMITATION OF LIABILITY
Open Pay''s total liability under this Agreement shall not exceed the total fees paid by the Merchant in the twelve (12) months preceding the claim. Open Pay shall not be liable for indirect, incidental, or consequential damages.

10. TERM AND TERMINATION
(a) This Agreement is effective upon signing and continues until terminated;
(b) Either party may terminate with 30 days'' written notice;
(c) Open Pay may immediately terminate if the Merchant breaches any material provision;
(d) Upon termination, all pending settlements will be processed within 30 business days.

11. DISPUTE RESOLUTION
Any disputes arising from this Agreement shall be resolved through:
(a) Good faith negotiation between the parties;
(b) If unresolved within 30 days, mediation under the rules of the Sri Lanka National Arbitration Centre;
(c) The courts of Sri Lanka shall have exclusive jurisdiction.

12. GOVERNING LAW
This Agreement shall be governed by the laws of the Democratic Socialist Republic of Sri Lanka.

13. AMENDMENTS
Open Pay reserves the right to amend this Agreement with 30 days'' prior notice. Continued use of the Service constitutes acceptance of amendments.

By signing below, the Merchant acknowledges having read, understood, and agreed to all terms of this Agreement.',
    TRUE
);

-- Seed default privacy policy (version 1)
INSERT INTO legal_documents (type, version, title, content, is_active) VALUES (
    'privacy_policy',
    1,
    'Open Pay — Privacy Policy',
    'OPEN PAY — PRIVACY POLICY

Last Updated: March 2026

Open Lanka Payment (Pvt) Ltd ("Open Pay", "we", "us", "our") is committed to protecting the privacy of our merchants, their customers, and website visitors. This Privacy Policy explains how we collect, use, disclose, and safeguard your information.

1. INFORMATION WE COLLECT

1.1 Merchant Information
When you register as a merchant, we collect:
(a) Business registration details and identification documents;
(b) Director and beneficial owner information (names, identification numbers, dates of birth);
(c) Banking and financial information for settlement purposes;
(d) Contact information (email, phone, address);
(e) Transaction history and payment processing data.

1.2 Customer Information
When customers make payments through our gateway, we may collect:
(a) Cryptocurrency wallet addresses;
(b) Transaction amounts and timestamps;
(c) Email addresses (if provided for receipts);
(d) IP addresses and device information for fraud prevention.

1.3 Automatically Collected Information
(a) Log data (IP address, browser type, access times);
(b) Device information and unique identifiers;
(c) Usage patterns and analytics data.

2. HOW WE USE YOUR INFORMATION
We use collected information to:
(a) Process payments and settlements;
(b) Verify merchant identity (KYC/AML compliance);
(c) Prevent fraud and unauthorized transactions;
(d) Communicate service updates and important notices;
(e) Comply with legal and regulatory requirements;
(f) Improve our services and user experience;
(g) Generate anonymized analytics and reporting.

3. LEGAL BASIS FOR PROCESSING
We process personal data under the following legal bases as defined by the Personal Data Protection Act No. 9 of 2022:
(a) Contractual necessity — to fulfill our service agreement with merchants;
(b) Legal obligation — to comply with AML/KYC regulations;
(c) Legitimate interests — for fraud prevention and service improvement;
(d) Consent — where specifically obtained for marketing communications.

4. DATA SHARING AND DISCLOSURE
We may share information with:
(a) Banking partners for settlement processing;
(b) Blockchain networks for payment verification;
(c) Regulatory authorities when required by law;
(d) Service providers who assist in our operations (under strict confidentiality agreements);
(e) Law enforcement agencies when legally compelled.

We do NOT sell personal information to third parties.

5. DATA SECURITY
We implement industry-standard security measures including:
(a) AES-256 encryption for data at rest;
(b) TLS 1.3 encryption for data in transit;
(c) Multi-factor authentication for account access;
(d) Regular security audits and penetration testing;
(e) Role-based access controls for internal systems;
(f) Secure data centers with physical access controls.

6. DATA RETENTION
(a) Active merchant data is retained for the duration of the business relationship;
(b) Transaction records are retained for 7 years as required by financial regulations;
(c) KYC documents are retained for 5 years after account closure;
(d) Marketing data is deleted upon withdrawal of consent.

7. YOUR RIGHTS
Under the Personal Data Protection Act, you have the right to:
(a) Access your personal data held by us;
(b) Request correction of inaccurate data;
(c) Request deletion of data (subject to legal retention requirements);
(d) Object to processing based on legitimate interests;
(e) Data portability — receive your data in a structured format;
(f) Withdraw consent for marketing communications at any time.

To exercise these rights, contact us at privacy@openpay.lk.

8. COOKIES AND TRACKING
Our platform uses:
(a) Essential cookies for authentication and security;
(b) Analytics cookies to understand usage patterns;
(c) You may disable non-essential cookies through your browser settings.

9. INTERNATIONAL DATA TRANSFERS
If data is transferred outside Sri Lanka, we ensure adequate protection through:
(a) Standard contractual clauses;
(b) Adequacy assessments of recipient countries;
(c) Compliance with PDPA cross-border transfer requirements.

10. CHILDREN''S PRIVACY
Our services are not directed to individuals under 18 years of age. We do not knowingly collect personal data from minors.

11. CHANGES TO THIS POLICY
We may update this Privacy Policy periodically. Material changes will be notified via email or dashboard notification at least 30 days before taking effect.

12. CONTACT US
For privacy-related inquiries:
Email: privacy@openpay.lk
Address: Open Lanka Payment (Pvt) Ltd, Colombo, Sri Lanka

13. GOVERNING LAW
This Privacy Policy is governed by the laws of the Democratic Socialist Republic of Sri Lanka, including the Personal Data Protection Act No. 9 of 2022.',
    TRUE
);
