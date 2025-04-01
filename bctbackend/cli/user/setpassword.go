package user

import (
	"bctbackend/database"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"fmt"

	_ "modernc.org/sqlite"
)

func SetPassword(databasePath string, userId models.Id, password string) error {
	db, err := database.ConnectToDatabase(databasePath)

	if err != nil {
		return err
	}

	defer db.Close()

	err = queries.UpdateUserPassword(db, userId, password)

	if err != nil {
		return err
	}

	fmt.Println("Password updated successfully")
	return nil
}
