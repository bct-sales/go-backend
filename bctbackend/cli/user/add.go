package user

import (
	"bctbackend/database"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"errors"
	"fmt"

	_ "modernc.org/sqlite"
)

func AddUser(databasePath string, userId models.Id, role string, password string) (r_err error) {
	db, err := database.OpenDatabase(databasePath)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	defer func() { r_err = errors.Join(r_err, db.Close()) }()

	roleId, err := models.ParseRole(role)
	if err != nil {
		return fmt.Errorf("invalid role %w; should be admin, seller or cashier", role)
	}

	timestamp := models.Now()
	var lastActivity *models.Timestamp = nil

	if err := queries.AddUserWithId(db, userId, roleId, timestamp, lastActivity, password); err != nil {
		var userIdAlreadyInUseError *queries.UserIdAlreadyInUseError
		if errors.As(err, &userIdAlreadyInUseError) {
			return fmt.Errorf("user ID %d is already in use", userId)
		}

		return err
	}

	fmt.Println("User added successfully")
	return nil
}
