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

func AddUser(databasePath string, userId models.Id, role string, password string) (err error) {
	db, err := database.ConnectToDatabase(databasePath)
	if err != nil {
		err = fmt.Errorf("failed to connect to database: %v", err)
		return err
	}

	defer func() { err = errors.Join(err, db.Close()) }()

	roleId, err := models.ParseRole(role)
	if err != nil {
		err = fmt.Errorf("invalid role %v; should be admin, seller or cashier", role)
		return err
	}

	timestamp := time.Now().Unix()
	var lastActivity *models.Timestamp = nil

	if err := queries.AddUserWithId(db, userId, roleId, timestamp, lastActivity, password); err != nil {
		var userIdAlreadyInUseError *queries.UserIdAlreadyInUseError
		if errors.As(err, &userIdAlreadyInUseError) {
			err = fmt.Errorf("user ID %d is already in use", userId)
			return err
		}

		return err
	}

	fmt.Println("User added successfully")
	err = nil
	return err
}
