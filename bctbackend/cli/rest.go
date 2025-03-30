package cli

import (
	database "bctbackend/database"
	rest "bctbackend/rest"
	"errors"

	_ "modernc.org/sqlite"
)

func startRestService(databasePath string) error {
	db, err := database.ConnectToDatabase(databasePath)

	if err != nil {
		return err
	}

	defer func() { err = errors.Join(err, db.Close()) }()

	if err = rest.StartRestService(db); err != nil {
		return err
	}

	err = nil
	return err
}
