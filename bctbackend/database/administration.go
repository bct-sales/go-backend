package database

import (
	models "bctbackend/database/models"
	defs "bctbackend/defs"
	"database/sql"
	"errors"
	"fmt"
	"log"
)

type DatabaseConnectionError struct {
	Context string
	Path    string
	Err     error
}

func (e *DatabaseConnectionError) Error() string {
	return fmt.Sprintf("failed to connect to database at %s while %s: %v", e.Path, e.Context, e.Err)
}

func (e *DatabaseConnectionError) Unwrap() error {
	return e.Err
}

func ConnectToDatabase(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path)

	if err != nil {
		return nil, &DatabaseConnectionError{Path: path, Err: err, Context: "opening database"}
	}

	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return nil, &DatabaseConnectionError{Path: path, Err: err, Context: "enabling foreign keys"}
	}

	return db, nil
}

func ResetDatabase(db *sql.DB) error {
	if err := removeAllTables(db); err != nil {
		return err
	}

	InitializeDatabase(db)

	return nil
}

func InitializeDatabase(db *sql.DB) error {
	if err := createTables(db); err != nil {
		return fmt.Errorf("failed to create tables: %v", err)
	}

	if err := createViews(db); err != nil {
		return fmt.Errorf("failed to create views: %v", err)
	}

	if err := populateTables(db); err != nil {
		return fmt.Errorf("failed to populate tables: %v", err)
	}

	return nil
}

func removeAllTables(db *sql.DB) error {
	tables := []string{"sessions", "sale_items", "sales", "items", "item_categories", "users", "roles"}

	for _, table := range tables {
		if _, err := db.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s", table)); err != nil {
			log.Fatalf("failed to drop table %s: %v", table, err)
		}
	}

	return nil
}

func createTables(db *sql.DB) error {
	if err := createRoleTable(db); err != nil {
		return err
	}

	if err := createUserTable(db); err != nil {
		return err
	}

	if err := createItemCategoryTable(db); err != nil {
		return err
	}

	if err := createItemTable(db); err != nil {
		return err
	}

	if err := createSaleTable(db); err != nil {
		return err
	}

	if err := createSaleItemsTable(db); err != nil {
		return err
	}

	if err := createSessionTable(db); err != nil {
		return err
	}

	return nil
}

func createRoleTable(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE roles (
			role_id             INTEGER NOT NULL,
			name                TEXT NOT NULL UNIQUE,

			PRIMARY KEY (role_id)
		)
	`)

	return err
}

func createUserTable(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE users (
			user_id             INTEGER NOT NULL,
			role_id             INTEGER NOT NULL,
			timestamp           INTEGER NOT NULL,
			password            TEXT NOT NULL,

			PRIMARY KEY (user_id),
			FOREIGN KEY (role_id) REFERENCES roles (role_id)
		);
	`)

	return err
}

func createItemCategoryTable(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE item_categories (
			item_category_id    INTEGER NOT NULL,
			name                TEXT NOT NULL UNIQUE,

			PRIMARY KEY (item_category_id)
		)
	`)

	return err
}

func createItemTable(db *sql.DB) error {
	_, err := db.Exec(`
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

	return err
}

func createSaleTable(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE sales (
			sale_id             INTEGER NOT NULL,
			cashier_id          INTEGER NOT NULL,
			timestamp           INTEGER NOT NULL,

			PRIMARY KEY (sale_id),
			FOREIGN KEY (cashier_id) REFERENCES users (user_id)
		)
	`)

	return err
}

func createSaleItemsTable(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE sale_items (
			sale_id             INTEGER NOT NULL,
			item_id             INTEGER NOT NULL,

			PRIMARY KEY (sale_id, item_id),
			FOREIGN KEY (sale_id) REFERENCES sales (sale_id),
			FOREIGN KEY (item_id) REFERENCES items (item_id)
		)
	`)

	return err
}

func createSessionTable(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE sessions (
			session_id          TEXT NOT NULL,
			user_id             INTEGER NOT NULL,

			PRIMARY KEY (session_id),
			FOREIGN KEY (user_id) REFERENCES users (user_id)
		)
	`)

	return err
}

func createViews(db *sql.DB) error {
	err := errors.Join(
		createCategoryCountsView(db),
	)

	return err
}

func createCategoryCountsView(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE VIEW item_category_counts AS
		SELECT
			item_categories.item_category_id as item_category_id,
			item_categories.name as item_category_name,
			COUNT(items.item_id) AS count
		FROM item_categories
		LEFT JOIN items ON item_categories.item_category_id = items.item_category_id
		GROUP BY item_categories.item_category_id
	`)

	return err
}

func populateTables(db *sql.DB) error {
	if err := populateRoleTable(db); err != nil {
		return err
	}

	if err := populateItemCategoryTable(db); err != nil {
		return err
	}

	return nil
}

func populateRoleTable(db *sql.DB) error {
	_, err := db.Exec(`
			INSERT INTO roles (role_id, name)
			VALUES
				($1, $2),
				($3, $4),
				($5, $6)
		`,
		models.AdminRoleId,
		models.AdminName,
		models.SellerRoleId,
		models.SellerName,
		models.CashierRoleId,
		models.CashierName,
	)

	if err != nil {
		return fmt.Errorf("failed to populate roles: %v", err)
	}

	return nil
}

func populateItemCategoryTable(db *sql.DB) error {
	for _, categoryId := range defs.ListCategories() {
		categoryName, err := defs.NameOfCategory(categoryId)

		if err != nil {
			return fmt.Errorf("failed to get category name: %v", err)
		}

		_, err = db.Exec(
			`
				INSERT INTO item_categories (item_category_id, name)
				VALUES ($1, $2)
			`,
			categoryId,
			categoryName,
		)

		if err != nil {
			return fmt.Errorf("failed to populate item categories: %v", err)
		}
	}

	return nil
}
