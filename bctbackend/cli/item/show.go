package item

import (
	"bctbackend/cli/formatting"
	"bctbackend/database"
	"bctbackend/database/models"
	"errors"

	_ "modernc.org/sqlite"
)

func ShowItem(databasePath string, itemId models.Id) (r_err error) {
	db, err := database.ConnectToDatabase(databasePath)
	if err != nil {
		return err
	}

	defer func() { r_err = errors.Join(r_err, db.Close()) }()

	err = formatting.PrintItem(db, itemId)
	if err != nil {
		return err
	}

	return nil
}
