package csv

import (
	"bctbackend/database"
	dbcsv "bctbackend/database/csv"
	"bctbackend/database/models"
	"bctbackend/database/queries"
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

	items := []*models.Item{}
	if err := queries.GetItems(db, queries.CollectTo(&items)); err != nil {
		return err
	}

	err = dbcsv.OutputItems(items, os.Stdout)
	if err != nil {
		return err
	}

	return nil
}
