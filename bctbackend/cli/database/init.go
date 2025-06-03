package database

import (
	database "bctbackend/database"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"errors"
	"fmt"
)

func InitializeDatabase(databasePath string, addCategories bool) (r_err error) {
	db, err := database.CreateDatabase(databasePath)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	defer func() { r_err = errors.Join(r_err, db.Close()) }()

	if err := database.InitializeDatabase(db); err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}

	if addCategories {
		if err := GenerateDefaultCategories(func(id models.Id, name string) error { return queries.AddCategoryWithId(db, id, name) }); err != nil {
			return fmt.Errorf("failed to add default categories: %w", err)
		}
	}

	fmt.Println("Database initialized successfully")
	return nil
}
