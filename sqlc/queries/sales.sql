-- name: AddSale :one
INSERT INTO sales(cashier_id, transaction_time)
VALUES (?, ?)
RETURNING sale_id;

-- name: AddSaleItem :exec
INSERT INTO sale_items(sale_id, item_id)
VALUES (?, ?);

-- name: GetSales :many
SELECT
    sales.sale_id,
    sales.cashier_id,
    sales.transaction_time,
    COUNT(sale_items.item_id) AS item_count,
    CAST(SUM(items.price_in_cents) AS INTEGER) AS total_price
FROM
    sales
INNER JOIN
    sale_items ON sales.sale_id = sale_items.sale_id
INNER JOIN
    items ON sale_items.item_id = items.item_id
WHERE
    sales.sale_id >= COALESCE(sqlc.narg('minimal_id'), sales.sale_id)
GROUP BY
    sales.sale_id
ORDER BY
    sales.sale_id ASC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: GetSalesChronologically :many
SELECT
    sales.sale_id,
    sales.cashier_id,
    sales.transaction_time,
    COUNT(sale_items.item_id) AS item_count,
    CAST(SUM(items.price_in_cents) AS INTEGER) AS total_price
FROM
    sales
INNER JOIN
    sale_items ON sales.sale_id = sale_items.sale_id
INNER JOIN
    items ON sale_items.item_id = items.item_id
WHERE
    sales.sale_id >= COALESCE(sqlc.narg('minimal_id'), sales.sale_id)
GROUP BY
    sales.sale_id
ORDER BY
    sales.transaction_time ASC, sales.sale_id ASC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: GetSalesAntiChronologically :many
SELECT
    sales.sale_id,
    sales.cashier_id,
    sales.transaction_time,
    COUNT(sale_items.item_id) AS item_count,
    CAST(SUM(items.price_in_cents) AS INTEGER) AS total_price
FROM
    sales
INNER JOIN
    sale_items ON sales.sale_id = sale_items.sale_id
INNER JOIN
    items ON sale_items.item_id = items.item_id
WHERE
    sales.sale_id >= COALESCE(sqlc.narg('minimal_id'), sales.sale_id)
GROUP BY
    sales.sale_id
ORDER BY
    sales.transaction_time DESC, sales.sale_id DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: GetSaleWithID :one
SELECT
    cashier_id,
    transaction_time
FROM
    sales
WHERE
    sale_id = ?;

-- name: SaleWithIDExists :one
SELECT 1
FROM
    sales
WHERE
    sale_id = ?;

-- name: GetSaleItems :many
SELECT
    i.item_id,
    i.added_at,
    i.description,
    i.price_in_cents,
    i.item_category_id,
    i.seller_id,
    i.donation,
    i.charity,
    i.frozen,
    i.large
FROM
    sale_items si
INNER JOIN
    items i ON si.item_id = i.item_id
WHERE
    si.sale_id = ?;

-- name: RemoveSaleItems :exec
DELETE FROM sale_items
WHERE
    sale_id = ?;

-- name: RemoveSale :exec
DELETE FROM sales
WHERE
    sale_id = ?;

-- name: RemoveAllSaleItems :exec
DELETE FROM sale_items;

-- name: RemoveAllSales :exec
DELETE FROM sales;

-- name: GetSoldItems :many
SELECT DISTINCT
    i.item_id,
    i.added_at,
    i.description,
    i.price_in_cents,
    i.item_category_id,
    i.seller_id,
    i.donation,
    i.charity,
    i.large,
    i.frozen
FROM
    sale_items si
INNER JOIN
    items i ON si.item_id = i.item_id
INNER JOIN
    sales s ON si.sale_id = s.sale_id
ORDER BY
    s.transaction_time DESC, i.item_id ASC;

-- name: ListSoldItems :many
SELECT
    sale.sale_id,
    sale.cashier_id,
    sale.transaction_time,
    item.item_id,
    item.added_at,
    item.description,
    item.price_in_cents,
    item.item_category_id,
    item.seller_id,
    item.donation,
    item.charity,
    item.large
FROM
    sales sale
INNER JOIN
    sale_items sale_item ON sale.sale_id = sale_item.sale_id
INNER JOIN
    items item ON sale_item.item_id = item.item_id
ORDER BY
    sale.sale_id, item.item_id;

-- name: CountSoldItems :one
SELECT
    COUNT(si.item_id),
    COUNT(DISTINCT si.item_id)
FROM
    sale_items si;

-- name: HasAnyBeenSold :one
SELECT EXISTS(
    SELECT 1
    FROM
        items
    INNER JOIN
        sale_items ON items.item_id = sale_items.item_id
    WHERE
        items.item_id IN (sqlc.slice('item_ids'))
);

-- name: GetItemsSoldBy :many
SELECT
    i.item_id,
    i.added_at,
    i.description,
    i.price_in_cents,
    i.item_category_id,
    i.seller_id,
    i.donation,
    i.charity,
    i.frozen,
    i.hidden,
    i.large
FROM
    sale_items si
INNER JOIN
    items i ON si.item_id = i.item_id
INNER JOIN
    sales s ON si.sale_id = s.sale_id
WHERE
    s.cashier_id = ?
ORDER BY
    s.transaction_time DESC, i.item_id ASC;

-- name: GetSalesWithItem :many
SELECT sale_id
FROM sale_items
WHERE item_id = ?
ORDER BY sale_id ASC;

-- name: GetSalesWithCashier :many
SELECT cashier_id, sale_id, transaction_time
FROM sales
WHERE cashier_id = ?
ORDER BY transaction_time ASC, sale_id ASC;

-- name: GetCashierSalesChronological :many
SELECT
    sales.sale_id,
    sales.cashier_id,
    sales.transaction_time,
    COUNT(sale_items.item_id) AS item_count,
    CAST(SUM(items.price_in_cents) AS INTEGER) AS total_price
FROM
    sales
INNER JOIN
    sale_items ON sales.sale_id = sale_items.sale_id
INNER JOIN
    items ON sale_items.item_id = items.item_id
WHERE
    sales.cashier_id = sqlc.arg('cashier_id')
GROUP BY
    sales.sale_id
ORDER BY
    sales.transaction_time ASC, sales.sale_id ASC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: GetCashierSalesAntiChronological :many
SELECT
    sales.sale_id,
    sales.cashier_id,
    sales.transaction_time,
    COUNT(sale_items.item_id) AS item_count,
    CAST(SUM(items.price_in_cents) AS INTEGER) AS total_price
FROM
    sales
INNER JOIN
    sale_items ON sales.sale_id = sale_items.sale_id
INNER JOIN
    items ON sale_items.item_id = items.item_id
WHERE
    sales.cashier_id = sqlc.arg('cashier_id')
GROUP BY
    sales.sale_id
ORDER BY
    sales.transaction_time DESC, sales.sale_id DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: CountSales :one
SELECT COUNT(*)
FROM sales;

-- name: CountCashierSales :one
SELECT COUNT(*)
FROM sales
WHERE cashier_id = ?;

-- name: GetTotalSalesValue :one
SELECT
    CAST(COALESCE(SUM(items.price_in_cents), 0) AS INTEGER) AS total
FROM
    sales
INNER JOIN
    sale_items ON sales.sale_id = sale_items.sale_id
INNER JOIN
    items ON sale_items.item_id = items.item_id;

-- name: GetMultiplySoldItems :many
SELECT
    item.item_id,
    item.added_at,
    item.description,
    item.price_in_cents,
    item.item_category_id,
    item.seller_id,
    item.donation,
    item.charity,
    item.large,
    item.frozen,
    sale.sale_id,
    sale.cashier_id,
    sale.transaction_time
FROM
    items item
INNER JOIN
    item_categories category ON item.item_category_id = category.item_category_id
INNER JOIN
    sale_items sale_item ON item.item_id = sale_item.item_id
INNER JOIN
    sales sale ON sale_item.sale_id = sale.sale_id
WHERE
    (
        SELECT COUNT(*)
        FROM sale_items si
        WHERE si.item_id = item.item_id
    ) > 1
ORDER BY
    item.item_id, sale.sale_id;

-- name: GetSaleItemInformation :one
SELECT seller_id, description, price_in_cents, item_category_id, large, COUNT(si.sale_id)
FROM items i LEFT JOIN sale_items si ON i.item_id = si.item_id
WHERE i.item_id = ?
GROUP BY i.item_id;

-- name: GetSalesOverview :many
SELECT
    item_categories.item_category_id,
    item_categories.name,
    CAST(SUM(CASE WHEN sale_items.item_id IS NOT NULL THEN items.price_in_cents ELSE 0 END) AS INTEGER) AS total
FROM
    item_categories
LEFT JOIN
    items ON item_categories.item_category_id = items.item_category_id
LEFT JOIN
    sale_items ON sale_items.item_id = items.item_id
GROUP BY
    item_categories.item_category_id
ORDER BY
    item_categories.item_category_id;
