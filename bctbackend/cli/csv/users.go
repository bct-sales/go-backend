package csv

import (
	"bctbackend/database"
	dbcsv "bctbackend/database/csv"
	"errors"
	"fmt"
	"os"

	_ "modernc.org/sqlite"
)

func ExportUsers(databasePath string) (r_err error) {
	db, err := database.OpenDatabase(databasePath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer func() { r_err = errors.Join(r_err, db.Close()) }()

	err = dbcsv.OutputUsers(db, os.Stdout)
	if err != nil {
		return fmt.Errorf("failed to output users: %w", err)
	}

	return nil
}
