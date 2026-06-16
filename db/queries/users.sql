-- name: CreateUser :exec
INSERT INTO users (
    id,
    username,
    display_name,
    bio,
    timezone,
    role,
    status,
    created_at,
    updated_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9);

-- name: GetUserByID :one
SELECT id, username, display_name, bio, timezone, role, status, created_at, updated_at
FROM users
WHERE id = $1;

-- name: GetUserByUsername :one
SELECT id, username, display_name, bio, timezone, role, status, created_at, updated_at
FROM users
WHERE username = $1;

-- name: UpdateUserProfile :execrows
UPDATE users
SET display_name = $2,
    bio = $3,
    timezone = $4,
    updated_at = $5
WHERE id = $1;
