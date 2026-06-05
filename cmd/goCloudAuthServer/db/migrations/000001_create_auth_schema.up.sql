-- go-cloud-k8s-auth: Authentication service schema
-- Schema: go_auth

CREATE SCHEMA IF NOT EXISTS go_auth;

SET search_path TO go_auth;

-- =============================================================================
-- Users: Core user table populated via OAuth provider login or admin creation
-- =============================================================================
CREATE TABLE IF NOT EXISTS go_auth.users (
    -- UUID v4 primary key, can be generated client-side or server-side
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    -- Auto-increment integer ID for legacy compatibility (JWT UserInfo.UserId)
    alternate_app_id  BIGSERIAL UNIQUE,
    email             TEXT UNIQUE NOT NULL,
    name              TEXT,
    avatar_url        TEXT,
    -- OAuth provider identifier: 'google', 'github', 'microsoft', 'local'
    provider          TEXT NOT NULL DEFAULT 'local',
    -- Provider-specific user ID (e.g., Google 'sub', GitHub user ID)
    provider_id       TEXT NOT NULL DEFAULT '',
    -- Phase 1 RBAC: simple role array ('user', 'admin', 'moderator', ...)
    roles             TEXT[] DEFAULT ARRAY['user'],
    is_active         BOOLEAN DEFAULT true,
    created_at        TIMESTAMPTZ DEFAULT NOW(),
    last_login_at     TIMESTAMPTZ,
    -- Ensure one user per provider+provider_id combination
    CONSTRAINT unique_provider UNIQUE (provider, provider_id)
);

-- =============================================================================
-- Groups: Named collections of users (personal, team, org)
-- =============================================================================
CREATE TABLE IF NOT EXISTS go_auth.groups (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name          TEXT NOT NULL,
    owner_id      UUID NOT NULL REFERENCES go_auth.users(id),
    -- Personal group is auto-created for each user
    is_personal   BOOLEAN DEFAULT false,
    created_at    TIMESTAMPTZ DEFAULT NOW()
);

-- =============================================================================
-- User <-> Group: Many-to-many with role within group
-- =============================================================================
CREATE TABLE IF NOT EXISTS go_auth.user_groups (
    user_id   UUID REFERENCES go_auth.users(id) ON DELETE CASCADE,
    group_id  UUID REFERENCES go_auth.groups(id) ON DELETE CASCADE,
    -- Role within the group: 'owner', 'admin', 'member'
    role      TEXT DEFAULT 'member',
    PRIMARY KEY (user_id, group_id)
);

-- =============================================================================
-- Indexes
-- =============================================================================
CREATE INDEX IF NOT EXISTS idx_users_email ON go_auth.users(email);
CREATE INDEX IF NOT EXISTS idx_users_provider ON go_auth.users(provider, provider_id);
CREATE INDEX IF NOT EXISTS idx_users_alternate_app_id ON go_auth.users(alternate_app_id);
CREATE INDEX IF NOT EXISTS idx_groups_owner ON go_auth.groups(owner_id);
CREATE INDEX IF NOT EXISTS idx_user_groups_user ON go_auth.user_groups(user_id);
CREATE INDEX IF NOT EXISTS idx_user_groups_group ON go_auth.user_groups(group_id);
