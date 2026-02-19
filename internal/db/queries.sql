-- name: SaveOTP :exec
INSERT INTO otps (email, otp, expires_at)
VALUES ($1, $2, $3);

-- name: GetOTP :one
SELECT * FROM otps
WHERE email = $1 AND expires_at > NOW();

-- name: DeleteOTP :exec
DELETE FROM otps WHERE email = $1;

-- name: CreateSession :one
INSERT INTO sessions (email, expires_at)
VALUES ($1, $2)
RETURNING *;

-- name: GetSession :one
SELECT * FROM sessions
WHERE token = $1 AND expires_at > NOW();
