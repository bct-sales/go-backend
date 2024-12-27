package cli

import (
	database "bctbackend/database"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"fmt"
	"log/slog"
	"time"

	_ "modernc.org/sqlite"
)

func resetDatabase(databasePath string) error {
	db, err := database.ConnectToDatabase(databasePath)

	if err != nil {
		return fmt.Errorf("failed to connect to database: %v", err)
	}

	defer db.Close()

	if err := database.ResetDatabase(db); err != nil {
		return fmt.Errorf("failed to reset database: %v", err)
	}

	fmt.Println("Database reset completed successfully")
	return nil
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

func resetDatabaseAndFillWithDummyData(databasePath string) error {
	db, err := database.ConnectToDatabase(databasePath)

	if err != nil {
		return fmt.Errorf("failed to connect to database: %v", err)
	}

	defer db.Close()

	if err := database.ResetDatabase(db); err != nil {
		return fmt.Errorf("failed to reset database: %v", err)
	}

	slog.Info("Adding admin user")
	queries.AddUserWithId(db, 1, models.AdminRoleId, time.Now().Unix(), "abc")

	for area := 1; area <= 12; area++ {
		for offset := 1; offset <= 10; offset++ {
			id := int64(area*100 + offset)
			slog.Info("Adding seller user", slog.Int64("id", id))
			queries.AddUserWithId(db, id, models.SellerRoleId, time.Now().Unix(), fmt.Sprintf("%d", id))
		}
	}

	return nil
}
