BEGIN;

CREATE TABLE refersh_tokens(
    id BIGSERIAL PRIMARY KEY,
    user_id BIGSERIAL NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash TEXT NOT NULL UNIQUE,
    family_id UUID NOT NULL,
    expeires_at TIMESTAMPTZ NOT NULL,
    used_at TIMESTAMPTZ ,
    revoked_at TIMESTAMPTZ,
    user_agent TEXT,
    created_at TIMESTAMPTZ DEFAULT now()
);

CREATE INDEX refresh_token_by_family_id ON refresh_tokens(family_id);

CREATE TABLE password_refresh_tokens(
    id BIGSERIAL PRIMARY KEY,
    user_id BIGSERIAL NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash TEXT,
    expires_at TIMESTAMPTZ NOT NULL,
    consumed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ
);

COMMIT;