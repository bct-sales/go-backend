PRAGMA foreign_keys = 1

CREATE TABLE roles (
    role_id             INTEGER NOT NULL,
    name                TEXT NOT NULL UNIQUE,

    PRIMARY KEY (role_id)
);

CREATE TABLE users (
    user_id             INTEGER NOT NULL,
    role_id             INTEGER NOT NULL,
    timestamp           INTEGER NOT NULL,
    password            TEXT NOT NULL,

    PRIMARY KEY (user_id),
    FOREIGN KEY (role_id) REFERENCES roles (role_id)
);

CREATE TABLE item_categories (
    item_category_id    INTEGER NOT NULL,
    name                TEXT NOT NULL UNIQUE,

    PRIMARY KEY (item_category_id)
);

CREATE TABLE items (
    item_id             INTEGER NOT NULL,
    timestamp           INTEGER NOT NULL,
    description         TEXT NOT NULL,
    price_in_cents      INTEGER NOT NULL,
    item_category_id    INTEGER NOT NULL,
    seller_id           INTEGER NOT NULL,
    donation            BOOLEAN NOT NULL,
    charity             BOOLEAN NOT NULL,

    PRIMARY KEY (item_id),
    FOREIGN KEY (seller_id) REFERENCES users (user_id),
    FOREIGN KEY (item_category_id) REFERENCES item_categories (item_category_id)
);

CREATE TABLE sales (
    sale_id             INTEGER NOT NULL,
    cashier_id          INTEGER NOT NULL,
    timestamp           INTEGER NOT NULL,

    PRIMARY KEY (sale_id),
    FOREIGN KEY (cashier_id) REFERENCES users (user_id)
);

CREATE TABLE sale_items (
    sale_id             INTEGER NOT NULL,
    item_id             INTEGER NOT NULL,

    PRIMARY KEY (sale_id, item_id),
    FOREIGN KEY (sale_id) REFERENCES sales (sale_id),
    FOREIGN KEY (item_id) REFERENCES items (item_id)
);

INSERT INTO roles (name)
VALUES
    ('admin'),
    ('seller'),
    ('cashier');

INSERT INTO item_categories (name)
VALUES
    ('Clothing 0-3 mos (50-56)'),
    ('Clothing 3-6 mos (56-62)'),
    ('Clothing 6-12 mos (68-80)'),
    ('Clothing 12-24 mos (86-92)'),
    ('Clothing 2-3 yrs (92-98)'),
    ('Clothing 4-6 yrs (104-116)'),
    ('Clothing 7-8 yrs (122-128)'),
    ('Clothing 9-10 yrs (128-140)'),
    ('Clothing 11-12 yrs (140-152)'),
    ('Shoes (infant to 12 yrs)'),
    ('Toys'),
    ('Baby/Child Equipment');
