CREATE TABLE IF NOT EXISTS secret (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
    type TEXT NOT NULL,
    metadata JSONB,
    data BYTEA,
    version INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS secrets_user_id_idx ON secret(user_id);
CREATE INDEX IF NOT EXISTS secrets_updated_at_idx ON secret(updated_at);