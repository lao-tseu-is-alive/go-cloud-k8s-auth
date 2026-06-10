package auth

// SQL query constants for the go_auth schema.
const (
	// --- Users ---

	// Matches an existing user by OAuth identity first, then by email
	// (providers verify email ownership, so a same-email login from another
	// provider re-links that account instead of violating users_email_key).
	// A plain ON CONFLICT can only target one unique constraint, hence the CTE.
	upsertUserByProvider = `
WITH target AS (
    SELECT id FROM go_auth.users
    WHERE (provider = $4 AND provider_id = $5) OR email = $1
    ORDER BY (provider = $4 AND provider_id = $5) DESC
    LIMIT 1
), updated AS (
    UPDATE go_auth.users u SET
        email = $1,
        name = COALESCE(NULLIF($2, ''), u.name),
        avatar_url = COALESCE(NULLIF($3, ''), u.avatar_url),
        provider = $4,
        provider_id = $5,
        last_login_at = NOW()
    FROM target t
    WHERE u.id = t.id
    RETURNING u.id, u.alternate_app_id, u.email, u.name, u.avatar_url, u.provider, u.provider_id, u.roles, u.is_active, u.created_at, u.last_login_at
), inserted AS (
    INSERT INTO go_auth.users (email, name, avatar_url, provider, provider_id, last_login_at)
    SELECT $1, $2, $3, $4, $5, NOW()
    WHERE NOT EXISTS (SELECT 1 FROM target)
    RETURNING id, alternate_app_id, email, name, avatar_url, provider, provider_id, roles, is_active, created_at, last_login_at
)
SELECT * FROM updated
UNION ALL
SELECT * FROM inserted;
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
