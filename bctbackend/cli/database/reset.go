package database

import (
	database "bctbackend/database"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"errors"
	"fmt"

	_ "modernc.org/sqlite"
)

func ResetDatabase(databasePath string, addCategories bool) (r_err error) {
	db, err := database.OpenDatabase(databasePath)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	defer func() { r_err = errors.Join(r_err, db.Close()) }()

	if err := database.ResetDatabase(db); err != nil {
		return fmt.Errorf("failed to reset database: %w", err)
	}

	if addCategories {
		GenerateDefaultCategories(func(id models.Id, name string) error {
			return queries.AddCategoryWithId(db, id, name)
		})
	}

	fmt.Println("Database reset completed successfully")
	return nil
}
