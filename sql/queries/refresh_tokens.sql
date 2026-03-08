-- name: CreateRefreshToken :one
WITH t AS (SELECT NOW() as ts)
INSERT INTO refresh_tokens (
  token, created_at, updated_at, user_id, expires_at, revoked_at
)
SELECT $1, ts, ts, $2, ts + INTERVAL '60 days', NULL
FROM t RETURNING *;

-- name: CheckRefreshToken :one
SELECT * FROM refresh_tokens WHERE token = $1;

-- name: RevokeRefreshToken :exec
WITH t as (SELECT NOW() as ts)
UPDATE refresh_tokens SET updated_at = t.ts, revoked_at = t.ts 
FROM t
WHERE token = $1;
