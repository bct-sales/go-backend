-- name: ListCategories :many
SELECT item_category_id, name
FROM item_categories
ORDER BY item_category_id;
