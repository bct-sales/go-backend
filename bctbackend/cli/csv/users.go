package csv

import (
	"bctbackend/database"
	dbcsv "bctbackend/database/csv"
	"os"

	_ "modernc.org/sqlite"
)

func ExportUsers(databasePath string) error {
	db, err := database.ConnectToDatabase(databasePath)
	if err != nil {
		return err
	}
	defer db.Close()

	err = dbcsv.OutputUsers(db, os.Stdout)
	if err != nil {
		return err
	}

	return nil
}
