-- name: AddSession :exec
INSERT INTO sessions (session_id, user_id, expiration_time)
VALUES (?, ?, ?);

-- name: GetSessionByID :one
SELECT
    user_id,
    expiration_time
FROM
    sessions
WHERE
    session_id = ?;

-- name: GetSessionData :one
SELECT
    users.user_id, role_id, expiration_time
FROM
    sessions
INNER JOIN
    users ON sessions.user_id = users.user_id
WHERE
    session_id = ? AND ? < expiration_time;

-- name: GetSessions :many
SELECT
    session_id,
    user_id,
    expiration_time
FROM
    sessions;

-- name: DeleteSession :execrows
DELETE FROM sessions
WHERE
    session_id = ?;

-- name: DeleteSessionWithUser :exec
DELETE FROM sessions
WHERE
    user_id = ?;

-- name: DeleteExpiredSessions :exec
DELETE FROM sessions
WHERE
    expiration_time < ?;

-- name: DeleteAllSessions :exec
DELETE FROM sessions;
