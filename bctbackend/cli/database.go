package cli

import (
	database "bctbackend/database"
	rest "bctbackend/rest"
	"fmt"

	_ "modernc.org/sqlite"
)

func startRestService(databasePath string) error {
	return rest.StartRestService(databasePath)
}

func resetDatabase(databasePath string) error {
	db, err := database.ConnectToDatabase(databasePath)

	if err != nil {
		return err
	}

	defer db.Close()

	if err := database.ResetDatabase(db); err != nil {
		fmt.Println(err)
		return err
	}

	fmt.Println("Database reset successfully")
	return nil
}

func backupDatabase(databasePath string, targetPath string) error {
	fmt.Printf("Backing up database to %s\n", targetPath)

	db, err := database.ConnectToDatabase(databasePath)

	if err != nil {
		return err
	}

	defer db.Close()

	if _, err = db.Exec("VACUUM INTO ?", targetPath); err != nil {
		return fmt.Errorf("failed to backup database %s to %s: %v", databasePath, targetPath, err)
	}

	fmt.Println("Database backup completed successfully")

	return nil
}
