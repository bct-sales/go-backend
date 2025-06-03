package cli

import (
	database "bctbackend/database"
	rest "bctbackend/rest"
	"errors"
	"fmt"
)

func startRestService(databasePath string) (r_err error) {
	db, err := database.OpenDatabase(databasePath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer func() { r_err = errors.Join(r_err, db.Close()) }()

	if err = rest.StartRestService(db); err != nil {
		return fmt.Errorf("failed to start REST service: %w", err)
	}

	return nil
}
