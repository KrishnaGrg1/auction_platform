-- name: CreateOTP :one
INSERT INTO otps (user_id, code, expires_at)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetValidOTPByUserId :one
SELECT * FROM otps
WHERE user_id      = $1
  AND used       = FALSE
  AND expires_at > NOW()
ORDER BY created_at DESC
LIMIT 1;

-- name: MarkOTPUsed :exec
UPDATE otps
SET used = TRUE
WHERE id = $1;

-- name: DeleteExpiredOTPs :exec
DELETE FROM otps
WHERE expires_at < NOW();
-- run this on a scheduler daily
-- keeps otps table clean

