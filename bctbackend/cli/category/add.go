package user

import (
	"bctbackend/database"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"errors"
	"fmt"
)

func AddCategory(databasePath string, id models.Id, name string) (r_err error) {
	db, err := database.OpenDatabase(databasePath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer func() { r_err = errors.Join(r_err, db.Close()) }()

	if err := queries.AddCategoryWithId(db, id, name); err != nil {
		return fmt.Errorf("failed to add category to database: %w", err)
	}

	fmt.Println("Category added successfully")

	return nil
}
