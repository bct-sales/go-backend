package cli

import (
	database "bctbackend/database"
	rest "bctbackend/rest"

	_ "modernc.org/sqlite"
)

func startRestService(databasePath string) error {
	db, err := database.ConnectToDatabase(databasePath)

	if err != nil {
		return err
	}

	defer db.Close()

	return rest.StartRestService(db)
}
