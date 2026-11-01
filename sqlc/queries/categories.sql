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
