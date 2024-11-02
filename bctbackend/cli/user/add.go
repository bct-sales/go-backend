package user

import (
	"bctbackend/database"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"time"

	_ "modernc.org/sqlite"
)

func AddUser(databasePath string, userId models.Id, role models.Id, password string) error {
	db, err := database.ConnectToDatabase(databasePath)

	if err != nil {
		return err
	}

	defer db.Close()

	timestamp := time.Now().Unix()

	if err := queries.AddUserWithId(db, userId, role, timestamp, password); err != nil {
		return err
	}

	return nil
}
