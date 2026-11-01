-- name: GetItemIDs :many
SELECT
    item_id
FROM
    items
ORDER BY
    item_id ASC;

-- name: GetSellerItems :many
SELECT
    item_id,
    added_at,
    description,
    price_in_cents,
    item_category_id,
    seller_id,
    donation,
    charity,
    frozen,
    hidden,
    large
FROM
    items
WHERE
    seller_id = ?
    AND hidden = COALESCE(sqlc.narg('hidden'), hidden)
ORDER BY
    added_at, item_id ASC;

-- name: GetItemsWithSaleCounts :many
SELECT
    i.item_id,
    added_at,
    description,
    price_in_cents,
    item_category_id,
    seller_id,
    donation,
    charity,
    frozen,
    hidden,
    large,
    CAST(COALESCE(COUNT(sale_items.sale_id), 0) AS INTEGER) AS sale_count
FROM
    items i LEFT JOIN sale_items ON i.item_id = sale_items.item_id
WHERE
    i.hidden = COALESCE(sqlc.narg('hidden'), i.hidden)
    AND i.seller_id = COALESCE(sqlc.narg('seller_id'), i.seller_id)
GROUP BY
    i.item_id
ORDER BY
    added_at, i.item_id ASC;

-- name: GetSellerItemsWithSaleCounts :many
SELECT
    items.item_id,
    added_at,
    description,
    price_in_cents,
    item_category_id,
    seller_id,
    donation,
    charity,
    frozen,
    hidden,
    large,
    CAST(COALESCE(COUNT(sale_items.sale_id), 0) AS INTEGER) AS sale_count
FROM
    items LEFT JOIN sale_items ON items.item_id = sale_items.item_id
WHERE
    seller_id = ? AND hidden = false
GROUP BY
    items.item_id
ORDER BY
    added_at, items.item_id ASC;

-- name: GetItemWithID :one
SELECT
    added_at,
    description,
    price_in_cents,
    item_category_id,
    seller_id,
    donation,
    charity,
    frozen,
    hidden,
    large
FROM
    items
WHERE
    item_id = ?;

-- name: ItemWithIDExists :one
SELECT 1
FROM
    items
WHERE
    item_id = ?;

-- name: CountItemsWithIDs :one
SELECT
    COUNT(item_id)
FROM
    items
WHERE
    item_id IN (sqlc.slice('item_ids'));

-- name: UpdateFreezeStatusOfItems :exec
UPDATE
    items
SET
    frozen = ?
WHERE
    item_id IN (sqlc.slice('item_ids'));

-- name: UpdateHiddenStatusOfItems :exec
UPDATE
    items
SET
    hidden = ?
WHERE
    item_id IN (sqlc.slice('item_ids'));

-- name: GetHiddenStatusOfItems :many
SELECT
    item_id,
    hidden
FROM
    items
WHERE
    item_id IN (sqlc.slice('item_ids'));

-- name: GetFrozenStatusOfItems :many
SELECT
    item_id,
    frozen
FROM
    items
WHERE
    item_id IN (sqlc.slice('item_ids'));

-- name: RemoveItemWithID :exec
DELETE FROM items
WHERE
    item_id = ?;

-- name: UpdateItem :exec
UPDATE
    items
SET
    added_at = COALESCE(sqlc.narg('added_at'), added_at),
    description = COALESCE(sqlc.narg('description'), description),
    price_in_cents = COALESCE(sqlc.narg('price_in_cents'), price_in_cents),
    item_category_id = COALESCE(sqlc.narg('item_category_id'), item_category_id),
    donation = COALESCE(sqlc.narg('donation'), donation),
    charity = COALESCE(sqlc.narg('charity'), charity),
    large = COALESCE(sqlc.narg('large'), large)
WHERE
    item_id = sqlc.arg('item_id');

-- name: AddItem :one
INSERT INTO
    items (added_at, description, price_in_cents, item_category_id, seller_id, donation, charity, frozen, hidden, large)
VALUES
    (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
RETURNING item_id;

-- name: DoesSellerHaveFrozenItems :one
SELECT
    1
FROM
    items
WHERE
    seller_id = ? AND items.frozen;

-- name: MoveItemsToNewSeller :exec
UPDATE
    items
SET
    seller_id = ?
WHERE
    seller_id = ?;

-- name: GetItemStatistics :one
SELECT
    COUNT(item_id),
    CAST(COALESCE(SUM(price_in_cents), 0) AS INTEGER)
FROM
    items
WHERE
    hidden = COALESCE(sqlc.narg('hidden'), hidden)
    AND description LIKE '%' || COALESCE(sqlc.narg('description'), description) || '%'
    AND item_category_id = COALESCE(sqlc.narg('item_category_id'), item_category_id);

-- name: GetItems :many
SELECT
    item_id,
    added_at,
    description,
    price_in_cents,
    item_category_id,
    seller_id,
    donation,
    charity,
    frozen,
    hidden,
    large
FROM
    items
WHERE
    hidden = COALESCE(sqlc.narg('hidden'), hidden)
    AND frozen = COALESCE(sqlc.narg('frozen'), frozen)
    AND item_category_id = COALESCE(sqlc.narg('item_category_id'), item_category_id)
    AND description LIKE COALESCE(sqlc.narg('description'), description)
    AND large = COALESCE(sqlc.narg('large'), large)
ORDER BY
    item_id ASC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: GetItemsWithIDs :many
SELECT
    item_id,
    added_at,
    description,
    price_in_cents,
    item_category_id,
    seller_id,
    donation,
    charity,
    frozen,
    hidden,
    large
FROM
    items
WHERE
    hidden = COALESCE(sqlc.narg('hidden'), hidden)
    AND frozen = COALESCE(sqlc.narg('frozen'), frozen)
    AND item_category_id = COALESCE(sqlc.narg('item_category_id'), item_category_id)
    AND description LIKE COALESCE(sqlc.narg('description'), description)
    AND large = COALESCE(sqlc.narg('large'), large)
    AND item_id IN (sqlc.slice('item_ids'))
ORDER BY
    item_id ASC;

-- name: ListItems :many
SELECT
    items.item_id,
    items.added_at,
    items.description,
    items.price_in_cents,
    items.item_category_id,
    items.seller_id,
    items.donation,
    items.charity,
    items.frozen,
    items.hidden,
    items.large,
    CAST(COALESCE(COUNT(sale_items.sale_id), 0) AS INTEGER) AS sale_count
FROM
    items
LEFT JOIN
    sale_items ON sale_items.item_id = items.item_id
WHERE
    items.large = COALESCE(sqlc.narg('large'), items.large)
    AND items.hidden = COALESCE(sqlc.narg('hidden'), items.hidden)
    AND items.item_category_id = COALESCE(sqlc.narg('item_category_id'), items.item_category_id)
    AND items.description LIKE COALESCE(sqlc.narg('description'), items.description)
GROUP BY
    items.item_id;
