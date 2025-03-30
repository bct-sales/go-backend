package cli

import (
	database "bctbackend/database"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"errors"
	"fmt"
	"log/slog"
	"time"

	_ "modernc.org/sqlite"
)

func resetDatabase(databasePath string) (err error) {
	db, err := database.ConnectToDatabase(databasePath)

	if err != nil {
		return fmt.Errorf("failed to connect to database: %v", err)
	}

	defer func() { err = errors.Join(err, db.Close()) }()

	if err := database.ResetDatabase(db); err != nil {
		return fmt.Errorf("failed to reset database: %v", err)
	}

	fmt.Println("Database reset completed successfully")
	err = nil
	return err
}

func backupDatabase(databasePath string, targetPath string) error {
	fmt.Printf("Backing up database to %s\n", targetPath)

	db, err := database.ConnectToDatabase(databasePath)

	if err != nil {
		return fmt.Errorf("failed to connect to database: %v", err)
	}

	defer db.Close()

	if _, err = db.Exec("VACUUM INTO ?", targetPath); err != nil {
		return fmt.Errorf("failed to backup database %s to %s: %v", databasePath, targetPath, err)
	}

	fmt.Println("Database backup completed successfully")
	return nil
}

func resetDatabaseAndFillWithDummyData(databasePath string) (err error) {
	db, err := database.ConnectToDatabase(databasePath)

	if err != nil {
		return fmt.Errorf("failed to connect to database: %v", err)
	}

	defer func() { err = errors.Join(err, db.Close()) }()

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
		for offset := 0; offset < 10; offset++ {
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

	return nil
}
