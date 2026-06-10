-- Browser SSO sessions: one row per long-lived session cookie.
-- The cookie value itself is never stored, only its SHA-256 hash.
CREATE TABLE IF NOT EXISTS go_auth.sessions
(
    id           UUID PRIMARY KEY     DEFAULT gen_random_uuid(),
    token_hash   TEXT UNIQUE NOT NULL, -- sha256 hex of the opaque cookie value
    user_id      UUID        NOT NULL REFERENCES go_auth.users (id) ON DELETE CASCADE,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at   TIMESTAMPTZ NOT NULL,
    last_seen_at TIMESTAMPTZ,
    user_agent   TEXT        NOT NULL DEFAULT '',
    ip           TEXT        NOT NULL DEFAULT '',
    revoked_at   TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_sessions_user ON go_auth.sessions (user_id);
CREATE INDEX IF NOT EXISTS idx_sessions_expires ON go_auth.sessions (expires_at);
