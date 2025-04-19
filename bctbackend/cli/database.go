package cli

import (
	database "bctbackend/database"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"bctbackend/defs"
	"errors"
	"fmt"
	"log/slog"
	"time"

	_ "modernc.org/sqlite"
)

func resetDatabase(databasePath string) (r_err error) {
	db, err := database.ConnectToDatabase(databasePath)

	if err != nil {
		return fmt.Errorf("failed to connect to database: %v", err)
	}

	defer func() { r_err = errors.Join(r_err, db.Close()) }()

	if err := database.ResetDatabase(db); err != nil {
		return fmt.Errorf("failed to reset database: %v", err)
	}

	fmt.Println("Database reset completed successfully")
	return nil
}

func backupDatabase(databasePath string, targetPath string) (r_err error) {
	fmt.Printf("Backing up database to %s\n", targetPath)

	db, err := database.ConnectToDatabase(databasePath)

	if err != nil {
		return fmt.Errorf("failed to connect to database: %v", err)
	}

	defer func() { r_err = errors.Join(r_err, db.Close()) }()

	if _, err = db.Exec("VACUUM INTO ?", targetPath); err != nil {
		return fmt.Errorf("failed to backup database %s to %s: %v", databasePath, targetPath, err)
	}

	fmt.Println("Database backup completed successfully")
	return nil
}

func resetDatabaseAndFillWithDummyData(databasePath string) (r_err error) {
	db, err := database.ConnectToDatabase(databasePath)

	if err != nil {
		return fmt.Errorf("failed to connect to database: %v", err)
	}

	defer func() { r_err = errors.Join(r_err, db.Close()) }()

	if err = database.ResetDatabase(db); err != nil {
		return fmt.Errorf("failed to reset database: %v", err)
	}

	slog.Info("Adding admin user")
	{
		id := models.NewId(1)
		role := models.AdminRoleId
		createdAt := time.Now().Unix()
		var lastActivity *models.Timestamp = nil
		password := "abc"

		if err = queries.AddUserWithId(db, id, role, createdAt, lastActivity, password); err != nil {
			return err
		}
	}

	for area := 1; area <= 12; area++ {
		for offset := 0; offset < 3; offset++ {
			id := int64(area*100 + offset)
			role := models.SellerRoleId
			createdAt := time.Now().Unix()
			var lastActivity *models.Timestamp = nil
			password := fmt.Sprintf("%d", id)

			slog.Info("Adding seller user", slog.Int64("id", id))
			if err = queries.AddUserWithId(db, id, role, createdAt, lastActivity, password); err != nil {
				return err
			}
		}
	}

	slog.Info("Adding some items")
	queries.AddItem(db, 1, "T-Shirt", 1000, defs.Clothing140_152, 100, false, false, false)
	queries.AddItem(db, 1, "Jeans", 1000, defs.Clothing140_152, 100, false, false, false)
	queries.AddItem(db, 1, "Nike sneakers", 1000, defs.Shoes, 100, false, false, false)

	return nil
}
