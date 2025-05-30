package database

import (
	models "bctbackend/database/models"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"os"
)

func fileExists(path string) (bool, error) {
	_, err := os.Stat(path)

	if err == nil {
		return true, nil
	}

	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}

	return false, err
}

// connectToDatabase opens a connection to the database at the specified path.
// If the database file does not exist, it is created.
func connectToDatabase(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", fmt.Sprintf("%s?_busy_timeout=500", path))
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	return db, nil
}

func enableForeignKeysConstraints(db *sql.DB) error {
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return fmt.Errorf("failed to enable foreign key constraints: %w", err)
	}

	return nil
}

func setJournalMode(db *sql.DB) error {
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		return fmt.Errorf("failed to set journal mode to WAL: %f", err)
	}

	return nil
}

func CreateDatabase(path string) (*sql.DB, error) {
	{
		slog.Debug("Ensuring no database file exists already", slog.String("path", path))
		exists, err := fileExists(path)

		if err != nil {
			slog.Debug("Error checking if database file exists", slog.String("path", path))
			return nil, err
		}

		if exists {
			slog.Debug("Database file already exists", slog.String("path", path))
			return nil, ErrDatabaseAlreadyExists
		}
	}

	slog.Debug("Creating database file", slog.String("path", path))
	db, err := connectToDatabase(path)
	if err != nil {
		return nil, fmt.Errorf("failed while creating database file: %w", err)
	}

	slog.Debug("Enabling foreign keys constraints", slog.String("path", path))
	if err := enableForeignKeysConstraints(db); err != nil {
		return nil, fmt.Errorf("failed to create database: %w", err)
	}

	slog.Debug("Setting journal mode", slog.String("path", path))
	if err := setJournalMode(db); err != nil {
		return nil, fmt.Errorf("failed to create database: %w", err)
	}

	return db, nil
}

func OpenDatabase(path string) (*sql.DB, error) {
	slog.Debug("Checking existence of database file", slog.String("path", path))
	if exists, err := fileExists(path); err != nil || !exists {
		slog.Debug("Database file not found", slog.String("path", path))
		return nil, fmt.Errorf("failed to check existence of database file: %w", err)
	}

	slog.Debug("Opening database file", slog.String("path", path))
	db, err := connectToDatabase(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	slog.Debug("Enabling foreign keys constraints", slog.String("path", path))
	if err := enableForeignKeysConstraints(db); err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	slog.Debug("Setting journal mode", slog.String("path", path))
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	slog.Debug("Connected to database", slog.String("path", path))
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
		return fmt.Errorf("failed to create tables: %w", err)
	}

	if err := createViews(db); err != nil {
		return fmt.Errorf("failed to create views: %w", err)
	}

	if err := populateTables(db); err != nil {
		return fmt.Errorf("failed to populate tables: %w", err)
	}

	return nil
}

func removeAllViews(db *sql.DB) error {
	views := []string{"visible_items", "hidden_items"}

	for _, view := range views {
		if err := dropView(db, view); err != nil {
			return fmt.Errorf("failed to drop view %s: %w", view, err)
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
	slog.Debug("Dropping table", slog.String("table", table))
	_, err := db.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s", table))

	if err != nil {
		return fmt.Errorf("failed to drop table %s: %w", table, err)
	}

	return nil
}

func dropView(db *sql.DB, view string) error {
	slog.Debug("Dropping view", slog.String("table", view))
	if _, err := db.Exec(fmt.Sprintf("DROP VIEW IF EXISTS %s", view)); err != nil {
		return fmt.Errorf("failed to drop view %s: %w", view, err)
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
	slog.Debug("Creating roles table")

	_, err := db.Exec(`
		CREATE TABLE roles (
			role_id             INTEGER NOT NULL,
			name                TEXT NOT NULL UNIQUE,

			PRIMARY KEY (role_id)
		)
	`)

	if err != nil {
		return fmt.Errorf("failed to create roles table: %w", err)
	}

	return nil
}

func createUserTable(db *sql.DB) error {
	slog.Debug("Creating users table")

	_, err := db.Exec(`
		CREATE TABLE users (
			user_id             INTEGER NOT NULL,
			role_id             INTEGER NOT NULL,
			created_at          INTEGER NOT NULL,
			last_activity       INTEGER,
			password            TEXT NOT NULL,

			PRIMARY KEY (user_id),
			CONSTRAINT users_foreign_key_role FOREIGN KEY (role_id) REFERENCES roles (role_id)
		);
	`)

	if err != nil {
		return fmt.Errorf("failed to create users table: %w", err)
	}

	return nil
}

func createItemCategoryTable(db *sql.DB) error {
	slog.Debug("Creating item categories table")

	_, err := db.Exec(`
		CREATE TABLE item_categories (
			item_category_id    INTEGER NOT NULL,
			name                TEXT NOT NULL UNIQUE,

			PRIMARY KEY (item_category_id)
		)
	`)

	if err != nil {
		return fmt.Errorf("failed to create item categories table: %w", err)
	}

	return nil
}

func createItemTable(db *sql.DB) error {
	slog.Debug("Creating items table")

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
			frozen              BOOLEAN NOT NULL,
			hidden              BOOLEAN NOT NULL,

			PRIMARY KEY (item_id),
			CONSTRAINT items_foreign_key_user FOREIGN KEY (seller_id) REFERENCES users (user_id),
			CONSTRAINT items_foreign_key_item_category FOREIGN KEY (item_category_id) REFERENCES item_categories (item_category_id)
		)
	`)

	if err != nil {
		return fmt.Errorf("failed to create items table: %w", err)
	}

	return nil
}

func createSaleTable(db *sql.DB) error {
	slog.Debug("Creating sales table")

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
		return fmt.Errorf("failed to create sales table: %w", err)
	}

	return nil
}

func createSaleItemsTable(db *sql.DB) error {
	slog.Debug("Creating sale items table")

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
		return fmt.Errorf("failed to create sale items table: %w", err)
	}

	return nil
}

func createSessionTable(db *sql.DB) error {
	slog.Debug("Creating sessions table")

	_, err := db.Exec(`
		CREATE TABLE sessions (
			session_id          TEXT NOT NULL,
			user_id             INTEGER NOT NULL,
			expiration_time     INTEGER NOT NULL,

			PRIMARY KEY (session_id),
			CONSTRAINT session_foreign_key_user FOREIGN KEY (user_id) REFERENCES users (user_id)
		)
	`)

	if err != nil {
		return fmt.Errorf("failed to create sessions table: %w", err)
	}

	return nil
}

func populateTables(db *sql.DB) error {
	if err := populateRoleTable(db); err != nil {
		return err
	}

	return nil
}

func populateRoleTable(db *sql.DB) error {
	slog.Debug("Populating roles table")

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
		return fmt.Errorf("failed to populate roles: %w", err)
	}

	return nil
}

func createViews(db *sql.DB) error {
	if err := createVisibleItemsView(db); err != nil {
		return fmt.Errorf("failed to create views: %w", err)
	}

	if err := createHiddenItemsView(db); err != nil {
		return fmt.Errorf("failed to create views: %w", err)
	}

	return nil
}

func createVisibleItemsView(db *sql.DB) error {
	slog.Debug("Creating visible items view")

	_, err := db.Exec(`
		CREATE VIEW visible_items AS
		SELECT *
		FROM items
		WHERE hidden = false
	`)

	if err != nil {
		return fmt.Errorf("failed to create visible_items view: %w", err)
	}

	return nil
}

func createHiddenItemsView(db *sql.DB) error {
	slog.Debug("Creating visible items view")

	_, err := db.Exec(`
		CREATE VIEW hidden_items AS
		SELECT *
		FROM items
		WHERE hidden = true
	`)

	if err != nil {
		return fmt.Errorf("failed to create hidden_items view: %w", err)
	}

	return nil
}
