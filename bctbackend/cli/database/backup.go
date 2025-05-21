package database

import (
	database "bctbackend/database"
	"errors"
	"fmt"

	_ "modernc.org/sqlite"
)

func BackupDatabase(databasePath string, targetPath string) (r_err error) {
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
