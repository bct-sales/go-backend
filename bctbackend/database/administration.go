package database

import (
	models "bctbackend/database/models"
	"bctbackend/defs"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
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
	if err := removeAllViews(db); err != nil {
		return err
	}

	if err := removeAllTables(db); err != nil {
		return err
	}

	return InitializeDatabase(db)
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

func removeAllViews(db *sql.DB) error {
	views := []string{"item_category_counts"}

	for _, view := range views {
		if err := dropView(db, view); err != nil {
			return fmt.Errorf("failed to drop view %s: %v", view, err)
		}
	}

	return nil
}

func removeAllTables(db *sql.DB) error {
	tables := []string{"sessions", "sale_items", "sales", "items", "item_categories", "users", "roles"}

	for _, table := range tables {
		if err := dropTable(db, table); err != nil {
			return err
		}
	}

	return nil
}

func dropTable(db *sql.DB, table string) error {
	slog.Info("Dropping table", slog.String("table", table))
	_, err := db.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s", table))

	if err != nil {
		return fmt.Errorf("failed to drop table %s: %v", table, err)
	}

	return nil
}

func dropView(db *sql.DB, view string) error {
	slog.Info("Dropping view", slog.String("table", view))
	_, err := db.Exec(fmt.Sprintf("DROP VIEW IF EXISTS %s", view))

	if err != nil {
		return fmt.Errorf("failed to drop view %s: %v", view, err)
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
	slog.Info("Creating roles table")

	_, err := db.Exec(`
		CREATE TABLE roles (
			role_id             INTEGER NOT NULL,
			name                TEXT NOT NULL UNIQUE,

			PRIMARY KEY (role_id)
		)
	`)

	if err != nil {
		return fmt.Errorf("failed to create roles table: %v", err)
	}

	return nil
}

func createUserTable(db *sql.DB) error {
	slog.Info("Creating users table")

	_, err := db.Exec(`
		CREATE TABLE users (
			user_id             INTEGER NOT NULL,
			role_id             INTEGER NOT NULL,
			created_at          INTEGER NOT NULL,
			password            TEXT NOT NULL,

			PRIMARY KEY (user_id),
			CONSTRAINT users_foreign_key_role FOREIGN KEY (role_id) REFERENCES roles (role_id)
		);
	`)

	if err != nil {
		return fmt.Errorf("failed to create users table: %v", err)
	}

	return nil
}

func createItemCategoryTable(db *sql.DB) error {
	slog.Info("Creating item categories table")

	_, err := db.Exec(`
		CREATE TABLE item_categories (
			item_category_id    INTEGER NOT NULL,
			name                TEXT NOT NULL UNIQUE,

			PRIMARY KEY (item_category_id)
		)
	`)

	if err != nil {
		return fmt.Errorf("failed to create item categories table: %v", err)
	}

	return nil
}

func createItemTable(db *sql.DB) error {
	slog.Info("Creating items table")

	_, err := db.Exec(`
		CREATE TABLE items (
			item_id             INTEGER NOT NULL,
			added_at            INTEGER NOT NULL,
			description         TEXT NOT NULL CHECK (LENGTH(description) > 0),
			price_in_cents      INTEGER NOT NULL CHECK (price_in_cents > 0),
			item_category_id    INTEGER NOT NULL,
			seller_id           INTEGER NOT NULL,
			donation            BOOLEAN NOT NULL,
			charity             BOOLEAN NOT NULL,

			PRIMARY KEY (item_id),
			CONSTRAINT items_foreign_key_user FOREIGN KEY (seller_id) REFERENCES users (user_id),
			CONSTRAINT items_foreign_key_item_category FOREIGN KEY (item_category_id) REFERENCES item_categories (item_category_id)
		)
	`)

	if err != nil {
		return fmt.Errorf("failed to create items table: %v", err
	}

	return nil
}

func createSaleTable(db *sql.DB) error {
	slog.Info("Creating sales table")

	_, err := db.Exec(`
		CREATE TABLE sales (
			sale_id             INTEGER NOT NULL,
			cashier_id          INTEGER NOT NULL,
			transaction_time    INTEGER NOT NULL,

			PRIMARY KEY (sale_id),
			CONSTRAINT sale_foreign_key_user FOREIGN KEY (cashier_id) REFERENCES users (user_id)
		)
	`)

	if err != nil {
		return fmt.Errorf("failed to create sales table: %v", err)
	}

	return nil
}

func createSaleItemsTable(db *sql.DB) error {
	slog.Info("Creating sale items table")

	_, err := db.Exec(`
		CREATE TABLE sale_items (
			sale_id             INTEGER NOT NULL,
			item_id             INTEGER NOT NULL,

			PRIMARY KEY (sale_id, item_id),
			CONSTRAINT sale_item_foreign_key_sale FOREIGN KEY (sale_id) REFERENCES sales (sale_id),
			CONSTRAINT sale_item_foreign_key_item FOREIGN KEY (item_id) REFERENCES items (item_id)
		)
	`)

	if err != nil {
		return fmt.Errorf("failed to create sale items table: %v", err)
	}

	return nil
}

func createSessionTable(db *sql.DB) error {
	slog.Info("Creating sessions table")

	_, err := db.Exec(`
		CREATE TABLE sessions (
			session_id          TEXT NOT NULL,
			user_id             INTEGER NOT NULL,
			expiration_time     INTEGER NOT NULL,

			PRIMARY KEY (session_id),
			CONSTRAINT session_foreign_key_user FOREIGN KEY (user_id) REFERENCES users (user_id)
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
	slog.Info("Creating item category counts view")

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
	slog.Info("Populating roles table")

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
	slog.Info("Populating item categories table")

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
