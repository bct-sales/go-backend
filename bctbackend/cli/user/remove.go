package user

import (
	"bctbackend/database"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"errors"
	"fmt"

	_ "modernc.org/sqlite"
)

func RemoveUser(databasePath string, userId models.Id) (r_err error) {
	db, err := database.ConnectToDatabase(databasePath)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	defer func() { r_err = errors.Join(r_err, db.Close()) }()

	if err = queries.RemoveUserWithId(db, userId); err != nil {
		return err
	}

	fmt.Println("User removed successfully")
	return nil
}
