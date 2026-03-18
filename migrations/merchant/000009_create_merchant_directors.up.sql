CREATE TABLE merchant_directors (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    merchant_id         UUID NOT NULL REFERENCES merchants(id) ON DELETE CASCADE,
    email               VARCHAR(255) NOT NULL,
    full_name           VARCHAR(255),
    date_of_birth       DATE,
    nic_passport_number VARCHAR(100),
    phone               VARCHAR(50),
    address             TEXT,
    document_object_key VARCHAR(500),
    document_filename   VARCHAR(255),
    verification_token  VARCHAR(100) UNIQUE NOT NULL,
    token_expires_at    TIMESTAMPTZ NOT NULL,
    status              VARCHAR(20) NOT NULL DEFAULT 'PENDING',
    consented_at        TIMESTAMPTZ,
    verified_at         TIMESTAMPTZ,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(merchant_id, email)
);

CREATE INDEX idx_directors_merchant ON merchant_directors(merchant_id);
CREATE INDEX idx_directors_token ON merchant_directors(verification_token);
