package user

import (
	"bctbackend/database"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"errors"
	"fmt"

	_ "modernc.org/sqlite"
)

func SetPassword(databasePath string, userId models.Id, password string) (r_err error) {
	db, err := database.OpenDatabase(databasePath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer func() { r_err = errors.Join(r_err, db.Close()) }()

	err = queries.UpdateUserPassword(db, userId, password)
	if err != nil {
		return fmt.Errorf("failed to update database: %w", err)
	}

	fmt.Println("Password updated successfully")
	return nil
}
