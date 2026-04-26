# Database

## `item_categories`

| Column | Type | Extra |
| ------ | ---- | ----- |
| `item_category_id` | `INTEGER NOT NULL` | Primary key |
| `name` | `TEXT NOT NULL UNIQUE` | |

## `items`

| Column | Type | Extra |
| ------ | ---- | ----- |
| `item_id` | `INTEGER NOT NULL` | Primary key |
| `added_at` | `INTEGER NOT NULL` | |
| `description` | `TEXT NOT NULL` | `CHECK (LENGTH(description) > 0)` |
| `price_in_cents` | `INTEGER NOT NULL` | `CHECK (price_in_cents > 0)` |
| `item_category_id` | `INTEGER NOT NULL` | Foreign key to `item_categories.item_category.id` |
| `seller_id` | `INTEGER NOT NULL` | Foreign key to `users.user_id` |
| `donation` | `BOOLEAN NOT NULL` | |
| `charity` | `BOOLEAN NOT NULL` | |
| `frozen` | `BOOLEAN NOT NULL` | |
| `hidden` | `BOOLEAN NOT NULL` | |
| `large` | `BOOLEAN NOT NULL` | |

## `sales`

| Column | Type | Extra |
| ------ | ---- | ----- |
| `sale_id` | `INTEGER NOT NULL` | Primary key |
| `cashier_id` | `INTEGER NOT NULL` | Foreigh key to `users.user_id` |
| `transaction_time` | `INTEGER NOT NULL` | |


## `sale_items`

| Column | Type | Extra |
| ------ | ---- | ----- |
| `sale_id` | `INTEGER NOT NULL` | Primary key, foreign key to `sales.sale_id` |
| `item_id` | `INTEGER NOT NULL` | Primary key, foreign key to `items.item_id` |

## `sessions`

| Column | Type | Extra |
| ------ | ---- | ----- |
| `session_id` | `TEXT NOT NULL` | Primary key |
| `user_id` | `INTEGER NOT NULL` | Foreign key to `users.user_id` |
| `expiration_time` | `INTEGER NOT NULL` | |
