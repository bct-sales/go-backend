package csv

import (
	"bctbackend/database"
	dbcsv "bctbackend/database/csv"
	"errors"
	"os"

	_ "modernc.org/sqlite"
)

func ExportItems(databasePath string) (r_err error) {
	db, err := database.ConnectToDatabase(databasePath)
	if err != nil {
		return err
	}
	defer func() { r_err = errors.Join(r_err, db.Close()) }()

	err = dbcsv.OutputItems(db, os.Stdout)
	if err != nil {
		return err
	}

	return nil
}
