-- Personal Access Tokens (PAT): long-lived opaque tokens for programmatic
-- access (MCP clients, scripts). The token value is never stored, only its
-- SHA-256 hash; the prefix is kept for display in management UIs.
CREATE TABLE IF NOT EXISTS go_auth.personal_access_tokens
(
    id           UUID PRIMARY KEY     DEFAULT gen_random_uuid(),
    user_id      UUID        NOT NULL REFERENCES go_auth.users (id) ON DELETE CASCADE,
    token_hash   TEXT UNIQUE NOT NULL, -- sha256 hex of the full "pat_..." string
    prefix       TEXT        NOT NULL, -- e.g. "pat_a1B2c3D4" for display
    name         TEXT        NOT NULL CHECK (length(btrim(name)) > 0 AND char_length(name) <= 100),
    scopes       TEXT[]      NOT NULL DEFAULT ARRAY ['notes:read','notes:write','notes:mcp'],
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at   TIMESTAMPTZ,          -- NULL = never expires
    last_used_at TIMESTAMPTZ,
    revoked_at   TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_pat_user ON go_auth.personal_access_tokens (user_id);
