package item

import (
	"bctbackend/cli/formatting"
	"bctbackend/database"
	"bctbackend/database/models"
	"errors"

	_ "modernc.org/sqlite"
)

func ShowItem(databasePath string, itemId models.Id) (err error) {
	db, err := database.ConnectToDatabase(databasePath)
	if err != nil {
		return err
	}

	defer func() { err = errors.Join(err, db.Close()) }()

	err = formatting.PrintItem(db, itemId)
	if err != nil {
		return err
	}

	err = nil
	return err
}
