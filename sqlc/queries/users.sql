-- name: GetUser :one
SELECT *
FROM users
WHERE user_id = ?;

-- name: AddUserWithID :exec
INSERT INTO users (user_id, role_id, created_at, last_activity, password)
VALUES (?, ?, ?, ?, ?)
