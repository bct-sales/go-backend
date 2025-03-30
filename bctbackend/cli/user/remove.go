package user

import (
	"bctbackend/database"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"errors"
	"fmt"

	_ "modernc.org/sqlite"
)

func RemoveUser(databasePath string, userId models.Id) error {
	db, err := database.ConnectToDatabase(databasePath)
	if err != nil {
		err = fmt.Errorf("failed to connect to database: %v", err)
		return err
	}

	defer func() { err = errors.Join(err, db.Close()) }()

	if err = queries.RemoveUserWithId(db, userId); err != nil {
		return err
	}

	fmt.Println("User removed successfully")
	err = nil
	return err
}
