package user

import (
	"bctbackend/database"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"errors"
	"fmt"
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
	var lastActivity *models.Timestamp = nil

	if err := queries.AddUserWithId(db, userId, role, timestamp, lastActivity, password); err != nil {
		{
			var userIdAlreadyInUseError *queries.UserIdAlreadyInUseError
			if errors.As(err, &userIdAlreadyInUseError) {
				fmt.Printf("User with id %d already exists\n", userId)
				return err
			}
		}

		


		return err
	}

	return nil
}
