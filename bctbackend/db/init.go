package db

import (
	models "bctbackend/db/models"
	"database/sql"
	"fmt"
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
			price_in_cents      INTEGER NOT NULL CHECK (price_in_cents > 0),
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
			FOREIGN KEY (item_id) REFERENCES items (item_id)
		)
	`)
}

func populateTables(db *sql.DB) {
	populateRoleTable(db)
	populateItemCategoryTable(db)
}

func populateRoleTable(db *sql.DB) {
	_, err := db.Exec(`
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

	if err != nil {
		panic(fmt.Errorf("failed to populate roles: %v", err))
	}
}

func populateItemCategoryTable(db *sql.DB) {
	_, err := db.Exec(`
		INSERT INTO item_categories (item_category_id, name)
		VALUES
			($1, 'Clothing 0-3 mos (50-56)'),
			($2, 'Clothing 3-6 mos (56-62)'),
			($3, 'Clothing 6-12 mos (68-80)'),
			($4, 'Clothing 12-24 mos (86-92)'),
			($5, 'Clothing 2-3 yrs (92-98)'),
			($6, 'Clothing 4-6 yrs (104-116)'),
			($7, 'Clothing 7-8 yrs (122-128)'),
			($8, 'Clothing 9-10 yrs (128-140)'),
			($9, 'Clothing 11-12 yrs (140-152)'),
			($10, 'Shoes (infant to 12 yrs)'),
			($11, 'Toys'),
			($12, 'Baby/Child Equipment')
		`,
		models.Clothing50_56,
		models.Clothing56_62,
		models.Clothing68_80,
		models.Clothing86_92,
		models.Clothing92_98,
		models.Clothing104_116,
		models.Clothing122_128,
		models.Clothing128_140,
		models.Clothing140_152,
		models.Shoes,
		models.Toys,
		models.BabyChildEquipment,
	)

	if err != nil {
		panic(fmt.Errorf("failed to populate item categories: %v", err))
	}
}
