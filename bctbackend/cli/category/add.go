package user

import (
	"bctbackend/database"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"errors"
	"fmt"

	_ "modernc.org/sqlite"
)

func AddCategory(databasePath string, id models.Id, name string) (r_err error) {
	db, err := database.ConnectToDatabase(databasePath)
	if err != nil {
		return err
	}
	defer func() { r_err = errors.Join(r_err, db.Close()) }()

	if err := queries.AddCategory(db, id, name); err != nil {
		return err
	}

	fmt.Println("Category added successfully")

	return nil
}
