package cli

import (
	database "bctbackend/database"
	rest "bctbackend/rest"
	"errors"

	_ "modernc.org/sqlite"
)

func startRestService(databasePath string) (r_err error) {
	db, err := database.OpenDatabase(databasePath)

	if err != nil {
		return err
	}

	defer func() { r_err = errors.Join(r_err, db.Close()) }()

	if err = rest.StartRestService(db); err != nil {
		return err
	}

	return nil
}
