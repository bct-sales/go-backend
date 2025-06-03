package user

import (
	"bctbackend/database"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"errors"
	"fmt"

	"github.com/urfave/cli/v2"
)

func AddUser(databasePath string, userId models.Id, role string, password string) (r_err error) {
	db, err := database.OpenDatabase(databasePath)
	if err != nil {
		return cli.Exit("Failed to connect to database: "+err.Error(), 1)
	}

	defer func() { r_err = errors.Join(r_err, db.Close()) }()

	roleId, err := models.ParseRole(role)
	if err != nil {
		return cli.Exit("Invalid role; should be admin, seller or cashier", 1)
	}

	timestamp := models.Now()
	var lastActivity *models.Timestamp = nil

	if err := queries.AddUserWithId(db, userId, roleId, timestamp, lastActivity, password); err != nil {
		return cli.Exit("Failed to add user: "+err.Error(), 1)
	}

	fmt.Println("User added successfully")
	return nil
}
