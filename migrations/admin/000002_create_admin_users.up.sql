-- Admin roles and permissions
CREATE TABLE admin_roles (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        VARCHAR(50) UNIQUE NOT NULL,
    description TEXT,
    permissions JSONB NOT NULL DEFAULT '[]',
    is_system   BOOLEAN DEFAULT FALSE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE admin_users (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email         VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    name          VARCHAR(255) NOT NULL,
    role_id       UUID NOT NULL REFERENCES admin_roles(id),
    is_active     BOOLEAN DEFAULT TRUE,
    last_login_at TIMESTAMPTZ,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_admin_users_email ON admin_users(email);
CREATE INDEX idx_admin_users_role ON admin_users(role_id);

-- Seed roles
INSERT INTO admin_roles (id, name, description, permissions, is_system) VALUES
(
    'a0000000-0000-0000-0000-000000000001',
    'SUPER_ADMIN',
    'Full platform access with all permissions',
    '["merchants:read","merchants:approve","merchants:reject","merchants:manage","withdrawals:read","withdrawals:approve","withdrawals:reject","withdrawals:complete","payments:read","treasury:read","treasury:manage","audit:read","subscriptions:read","notifications:read","system:manage"]',
    TRUE
),
(
    'a0000000-0000-0000-0000-000000000002',
    'ADMIN',
    'Standard admin with review and approval permissions',
    '["merchants:read","merchants:approve","merchants:reject","withdrawals:read","withdrawals:approve","withdrawals:reject","withdrawals:complete","payments:read","treasury:read","audit:read","subscriptions:read","notifications:read"]',
    TRUE
),
(
    'a0000000-0000-0000-0000-000000000003',
    'VIEWER',
    'Read-only access to platform data',
    '["merchants:read","withdrawals:read","payments:read","treasury:read","audit:read","subscriptions:read","notifications:read"]',
    TRUE
);

-- Seed default super admin (password: Admin@2024)
-- bcrypt hash of 'Admin@2024'
INSERT INTO admin_users (id, email, password_hash, name, role_id) VALUES
(
    'a0000000-0000-0000-0000-100000000001',
    'admin@openlankapay.lk',
    '$2a$10$J4z0DeGTDIDwPTIHl7shPe3LysG12Yl8sa6I9f5TRB57WErT04ECq',
    'Platform Admin',
    'a0000000-0000-0000-0000-000000000001'
);
