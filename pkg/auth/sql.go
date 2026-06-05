package auth

// SQL query constants for the go_auth schema.
const (
	// --- Users ---

	upsertUserByProvider = `
INSERT INTO go_auth.users (email, name, avatar_url, provider, provider_id, last_login_at)
VALUES ($1, $2, $3, $4, $5, NOW())
ON CONFLICT (provider, provider_id) DO UPDATE SET
    email = EXCLUDED.email,
    name = COALESCE(NULLIF(EXCLUDED.name, ''), go_auth.users.name),
    avatar_url = COALESCE(NULLIF(EXCLUDED.avatar_url, ''), go_auth.users.avatar_url),
    last_login_at = NOW()
RETURNING id, alternate_app_id, email, name, avatar_url, provider, provider_id, roles, is_active, created_at, last_login_at;
`

	getUserByID = `
SELECT id, alternate_app_id, email, name, avatar_url, provider, provider_id, roles, is_active, created_at, last_login_at
FROM go_auth.users
WHERE id = $1;
`

	getUserByAlternateAppID = `
SELECT id, alternate_app_id, email, name, avatar_url, provider, provider_id, roles, is_active, created_at, last_login_at
FROM go_auth.users
WHERE alternate_app_id = $1;
`

	getUserByExternalID = `
SELECT id, alternate_app_id, email, name, avatar_url, provider, provider_id, roles, is_active, created_at, last_login_at
FROM go_auth.users
WHERE alternate_app_id = $1;
`

	listUsers = `
SELECT id, alternate_app_id, email, name, is_active, created_at, last_login_at
FROM go_auth.users
WHERE ($3::boolean IS NULL OR is_active = $3)
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;
`

	createUser = `
INSERT INTO go_auth.users (email, name, avatar_url, provider, provider_id, roles, is_active)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id, alternate_app_id, email, name, avatar_url, provider, provider_id, roles, is_active, created_at, last_login_at;
`

	updateUser = `
UPDATE go_auth.users SET
    email = $2,
    name = $3,
    avatar_url = $4,
    roles = $5,
    is_active = $6
WHERE id = $1
RETURNING id, alternate_app_id, email, name, avatar_url, provider, provider_id, roles, is_active, created_at, last_login_at;
`

	deleteUser = `
DELETE FROM go_auth.users WHERE id = $1;
`

	countUsers = `
SELECT COUNT(*) FROM go_auth.users
WHERE ($1::boolean IS NULL OR is_active = $1);
`

	existUser = `SELECT COUNT(*) FROM go_auth.users WHERE id = $1;`

	updateLastLogin = `UPDATE go_auth.users SET last_login_at = NOW() WHERE id = $1;`

	// --- Groups ---

	getUserGroupIDs = `
SELECT ug.group_id
FROM go_auth.user_groups ug
WHERE ug.user_id = $1;
`

	// Alternate: get integer-based group IDs for JWT compatibility
	// We cast UUID to text and use a deterministic mapping via user_groups row number
	// For Phase 1, we use the user_groups join and return the alternate_app_id of group members
	// Actually, for JWT Groups []int, we need integer group IDs.
	// We'll use a helper that returns group alternate IDs or sequential ints.
	// For now, return group UUIDs and handle mapping in the service layer.

	getUserGroupIDsAsInts = `
SELECT DISTINCT ug.group_id
FROM go_auth.user_groups ug
WHERE ug.user_id = $1;
`
)
