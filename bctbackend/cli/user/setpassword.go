package user

import (
	"bctbackend/database"
	"bctbackend/database/models"
	"bctbackend/database/queries"

	_ "modernc.org/sqlite"
)

func SetPassword(databasePath string, userId models.Id, password string) error {
	db, err := database.ConnectToDatabase(databasePath)

	if err != nil {
		return err
	}

	defer db.Close()

	return queries.UpdateUserPassword(db, userId, password)
}
