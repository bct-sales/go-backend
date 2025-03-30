package csv

import (
	"bctbackend/database"
	dbcsv "bctbackend/database/csv"
	"errors"
	"os"

	_ "modernc.org/sqlite"
)

func ExportItems(databasePath string) error {
	db, err := database.ConnectToDatabase(databasePath)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, db.Close()) }()

	err = dbcsv.OutputItems(db, os.Stdout)
	if err != nil {
		return err
	}

	err = nil
	return err
}
