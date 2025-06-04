package common

import (
	"bctbackend/database"
	"database/sql"
	"fmt"

	"errors"
)

func WithOpenedDatabase(databasePath string, fn func(db *sql.DB) error) (r_err error) {
	db, err := database.OpenDatabase(databasePath)
	if err != nil {
		return fmt.Errorf("failed to open database %s: %w", databasePath, err)
	}
	defer func() { r_err = errors.Join(r_err, db.Close()) }()

	return fn(db)
}
