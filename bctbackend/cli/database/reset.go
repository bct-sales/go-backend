package database

import (
	database "bctbackend/database"
	"errors"
	"fmt"

	_ "modernc.org/sqlite"
)

func ResetDatabase(databasePath string) (r_err error) {
	db, err := database.ConnectToDatabase(databasePath)

	if err != nil {
		return fmt.Errorf("failed to connect to database: %v", err)
	}

	defer func() { r_err = errors.Join(r_err, db.Close()) }()

	if err := database.ResetDatabase(db); err != nil {
		return fmt.Errorf("failed to reset database: %v", err)
	}

	fmt.Println("Database reset completed successfully")
	return nil
}
