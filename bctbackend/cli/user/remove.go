package user

import (
	"bctbackend/database"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"errors"
	"fmt"

	"github.com/urfave/cli/v2"
	_ "modernc.org/sqlite"
)

func RemoveUser(databasePath string, userId models.Id) (r_err error) {
	db, err := database.OpenDatabase(databasePath)
	if err != nil {
		return cli.Exit("Failed to connect to database", 1)
	}

	defer func() { r_err = errors.Join(r_err, db.Close()) }()

	if err = queries.RemoveUserWithId(db, userId); err != nil {
		return err
	}

	fmt.Println("User removed successfully")
	return nil
}
