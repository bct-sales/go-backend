-- name: GetUser :one
SELECT *
FROM users
WHERE user_id = ?;

-- name: AddUserWithID :exec
INSERT INTO users (user_id, role_id, created_at, last_activity, password)
VALUES (?, ?, ?, ?, ?);

-- name: AddUser :one
INSERT INTO users (role_id, created_at, last_activity, password)
VALUES (?, ?, ?, ?)
RETURNING user_id;

-- name: UserWithIDExists :one
SELECT 1
FROM
    users
WHERE
    user_id = ?;

-- name: GetUsers :many
SELECT
    user_id,
    role_id,
    created_at,
    last_activity,
    password
FROM
    users;

-- name: GetUsersWithItemCount :many
SELECT
    users.user_id,
    users.role_id,
    users.created_at,
    users.last_activity,
    users.password,
    CAST(COALESCE(COUNT(i.item_id), 0) AS INTEGER) AS item_count
FROM
    users
LEFT JOIN
    items i ON users.user_id = i.seller_id AND i.hidden = COALESCE(sqlc.narg('hidden'), i.hidden)
GROUP BY
    users.user_id
ORDER BY
    users.user_id;

-- name: UpdateUserPassword :exec
UPDATE
    users
SET
    password = ?
WHERE
    user_id = ?;

-- name: RemoveUserWithID :exec
DELETE FROM users
WHERE
    user_id = ?;

-- name: UpdateLastActivity :exec
UPDATE
    users
SET
    last_activity = ?
WHERE
    user_id = ?;

-- name: CountSellerItems :one
SELECT COUNT(item_id)
FROM
    items
WHERE
    seller_id = ?
    AND frozen = COALESCE(sqlc.narg('frozen'), frozen)
    AND hidden = COALESCE(sqlc.narg('hidden'), hidden);

-- name: GetSellerTotalValueOfAllItems :one
SELECT CAST(COALESCE(SUM(price_in_cents), 0) AS INTEGER)
FROM
    items
WHERE
    seller_id = ?
    AND hidden = COALESCE(sqlc.narg('hidden'), hidden);

-- name: CollectPasswords :many
SELECT password FROM users;

-- name: AuthenticateUser :one
SELECT
    role_id,
    password
FROM
    users
WHERE
    user_id = ?;
