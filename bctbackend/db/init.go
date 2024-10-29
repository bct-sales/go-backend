package db

import (
	models "bctbackend/db/models"
	"database/sql"
)

func InitializeDatabase(db *sql.DB) {
	createTables(db)
	populateTables(db)
}

func createTables(db *sql.DB) {
	createRoleTable(db)
	createUserTable(db)
	createItemCategoryTable(db)
	createItemTable(db)
	createSaleTable(db)
	createSaleItemsTable(db)
}

func createRoleTable(db *sql.DB) {
	db.Exec(`
		CREATE TABLE roles (
			role_id             INTEGER NOT NULL,
			name                TEXT NOT NULL UNIQUE,

			PRIMARY KEY (role_id)
		)
	`)
}

func createUserTable(db *sql.DB) {
	db.Exec(`
		CREATE TABLE users (
			user_id             INTEGER NOT NULL,
			role_id             INTEGER NOT NULL,
			timestamp           INTEGER NOT NULL,
			password            TEXT NOT NULL,

			PRIMARY KEY (user_id),
			FOREIGN KEY (role_id) REFERENCES roles (role_id)
		);
	`)
}

func createItemCategoryTable(db *sql.DB) {
	db.Exec(`
		CREATE TABLE item_categories (
			item_category_id    INTEGER NOT NULL,
			name                TEXT NOT NULL UNIQUE,

			PRIMARY KEY (item_category_id)
		)
	`)
}

func createItemTable(db *sql.DB) {
	db.Exec(`
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
		)
	`)
}

func createSaleTable(db *sql.DB) {
	db.Exec(`
		CREATE TABLE sales (
			sale_id             INTEGER NOT NULL,
			cashier_id          INTEGER NOT NULL,
			timestamp           INTEGER NOT NULL,

			PRIMARY KEY (sale_id),
			FOREIGN KEY (cashier_id) REFERENCES users (user_id)
		)
	`)
}

func createSaleItemsTable(db *sql.DB) {
	db.Exec(`
		CREATE TABLE sale_items (
			sale_id             INTEGER NOT NULL,
			item_id             INTEGER NOT NULL,

			PRIMARY KEY (sale_id, item_id),
			FOREIGN KEY (sale_id) REFERENCES sales (sale_id),
			FOREIGN KEY (item_id) REFERENCES sales (item_id)
		)
	`)
}

func populateTables(db *sql.DB) {
	populateRoleTable(db)
	populateItemCategoryTable(db)
}

func populateRoleTable(db *sql.DB) {
	db.Exec(`
			INSERT INTO roles (role_id, name)
			VALUES
				($1, 'admin'),
				($2, 'seller'),
				($3, 'cashier')
		`,
		models.AdminRoleId,
		models.SellerRoleId,
		models.CashierRoleId,
	)
}

func populateItemCategoryTable(db *sql.DB) {
	db.Exec(`
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
			('Baby/Child Equipment')
	`)
}
