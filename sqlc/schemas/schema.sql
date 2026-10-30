CREATE TABLE roles (
    role_id             INTEGER NOT NULL,
    name                TEXT NOT NULL UNIQUE,

    PRIMARY KEY (role_id)
);

CREATE TABLE users (
    user_id             INTEGER NOT NULL,
    role_id             INTEGER NOT NULL,
    created_at          INTEGER NOT NULL,
    last_activity       INTEGER,
    password            TEXT NOT NULL,

    PRIMARY KEY (user_id),
    CONSTRAINT users_foreign_key_role FOREIGN KEY (role_id) REFERENCES roles (role_id)
);

CREATE TABLE item_categories (
    item_category_id    INTEGER NOT NULL,
    name                TEXT NOT NULL UNIQUE,

    PRIMARY KEY (item_category_id)
);

CREATE TABLE items (
    item_id             INTEGER NOT NULL,
    added_at            INTEGER NOT NULL,
    description         TEXT NOT NULL CHECK (LENGTH(description) > 0),
    price_in_cents      INTEGER NOT NULL CHECK (price_in_cents > 0),
    item_category_id    INTEGER NOT NULL,
    seller_id           INTEGER NOT NULL,
    donation            BOOLEAN NOT NULL,
    charity             BOOLEAN NOT NULL,
    frozen              BOOLEAN NOT NULL,
    hidden              BOOLEAN NOT NULL,
    large               BOOLEAN NOT NULL,

    PRIMARY KEY (item_id),
    CONSTRAINT items_foreign_key_user
                FOREIGN KEY (seller_id) REFERENCES users (user_id),
    CONSTRAINT items_foreign_key_item_category
                FOREIGN KEY (item_category_id) REFERENCES item_categories (item_category_id)
);

CREATE TABLE sales (
    sale_id             INTEGER NOT NULL,
    cashier_id          INTEGER NOT NULL,
    transaction_time    INTEGER NOT NULL,

    PRIMARY KEY (sale_id),
    CONSTRAINT sale_foreign_key_user FOREIGN KEY (cashier_id) REFERENCES users (user_id)
);

CREATE TABLE sale_items (
    sale_id             INTEGER NOT NULL,
    item_id             INTEGER NOT NULL,

    PRIMARY KEY (sale_id, item_id),
    CONSTRAINT sale_item_foreign_key_sale FOREIGN KEY (sale_id) REFERENCES sales (sale_id),
    CONSTRAINT sale_item_foreign_key_item FOREIGN KEY (item_id) REFERENCES items (item_id)
);

CREATE TABLE sessions (
    session_id          TEXT NOT NULL,
    user_id             INTEGER NOT NULL,
    expiration_time     INTEGER NOT NULL,

    PRIMARY KEY (session_id),
    CONSTRAINT session_foreign_key_user FOREIGN KEY (user_id) REFERENCES users (user_id)
);

