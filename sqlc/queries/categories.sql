-- name: ListCategories :many
SELECT item_category_id, name
FROM item_categories
ORDER BY item_category_id;

-- name: CountItemsPerCategory :many
SELECT
    item_categories.item_category_id, COUNT(items.item_id)
FROM
    item_categories
LEFT JOIN
    items ON item_categories.item_category_id = items.item_category_id
GROUP BY
    item_categories.item_category_id;

-- name: CountItemsPerCategoryByHidden :many
SELECT
    item_categories.item_category_id, COUNT(items.item_id)
FROM
    item_categories
LEFT JOIN
    items ON item_categories.item_category_id = items.item_category_id AND items.hidden = ?
GROUP BY
    item_categories.item_category_id;

-- name: AddCategory :one
INSERT INTO
    item_categories(name)
VALUES
    (?)
RETURNING item_category_id;

-- name: CategoryWithIDExists :one
SELECT 1
FROM
    item_categories
WHERE
    item_category_id = ?;

-- name: AddCategoryWithID :exec
INSERT INTO
    item_categories(item_category_id, name)
VALUES
    (?, ?);

-- name: CountSoldItemsPerCategory :many
SELECT
    item_categories.item_category_id, COUNT(sale_items.item_id)
FROM
    item_categories
LEFT JOIN
    items ON item_categories.item_category_id = items.item_category_id
LEFT JOIN
    sale_items ON sale_items.item_id = items.item_id
GROUP BY
    item_categories.item_category_id;

-- name: CategoryWithNameExists :one
SELECT 1
FROM
    item_categories
WHERE
    name = ?;

-- name: RenameCategory :exec
UPDATE
    item_categories
SET
    name = ?
WHERE
    item_category_id = ?;
