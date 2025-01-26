package user

import (
	"bctbackend/database"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"fmt"

	_ "modernc.org/sqlite"
)

func RemoveUser(databasePath string, userId models.Id) error {
	db, err := database.ConnectToDatabase(databasePath)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %v", err)
	}
	defer db.Close()

	err = queries.RemoveUserWithId(db, userId)
	if err != nil {
		return err
	}

	fmt.Println("User removed successfully")
	return nil
}
