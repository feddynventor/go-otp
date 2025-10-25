CREATE TABLE IF NOT EXISTS master_tokens (
    id VARCHAR(36) PRIMARY KEY,
    secret VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    issuer VARCHAR(255),
    account_name VARCHAR(255)
);

CREATE INDEX IF NOT EXISTS idx_master_tokens_active ON master_tokens(is_active);
CREATE INDEX IF NOT EXISTS idx_master_tokens_created_at ON master_tokens(created_at);
