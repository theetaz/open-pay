CREATE TABLE users (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    merchant_id     UUID NOT NULL REFERENCES merchants(id),
    email           VARCHAR(255) NOT NULL,
    password_hash   VARCHAR(255) NOT NULL,
    name            VARCHAR(255),
    role            VARCHAR(20) NOT NULL DEFAULT 'USER'
                    CHECK (role IN ('ADMIN', 'MANAGER', 'USER')),
    branch_id       UUID REFERENCES branches(id),
    is_active       BOOLEAN DEFAULT TRUE,
    last_login_at   TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    UNIQUE(merchant_id, email)
);

CREATE INDEX idx_users_merchant ON users(merchant_id);
CREATE INDEX idx_users_email ON users(email);
