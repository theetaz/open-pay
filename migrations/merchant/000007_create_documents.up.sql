CREATE TABLE documents (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    merchant_id   UUID NOT NULL REFERENCES merchants(id),
    category      VARCHAR(50) NOT NULL,
    filename      VARCHAR(500) NOT NULL,
    object_key    VARCHAR(1000) NOT NULL,
    content_type  VARCHAR(100) NOT NULL,
    file_size     BIGINT NOT NULL,
    status        VARCHAR(20) NOT NULL DEFAULT 'UPLOADED' CHECK (status IN ('UPLOADED','VERIFIED','REJECTED')),
    uploaded_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    verified_at   TIMESTAMPTZ,
    verified_by   UUID
);

CREATE INDEX idx_documents_merchant ON documents(merchant_id);
CREATE INDEX idx_documents_category ON documents(merchant_id, category);
