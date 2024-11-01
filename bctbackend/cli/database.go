package cli

import (
	database "bctbackend/database"
	rest "bctbackend/rest"
	"fmt"
	"log"

	"database/sql"

	"github.com/urfave/cli/v2"

	_ "modernc.org/sqlite"
)

func startRestService(context *cli.Context) error {
	return rest.StartRestService()
}

func resetDatabase(databasePath string) error {
	db, err := sql.Open("sqlite", databasePath)

	if err != nil {
		log.Fatalf("Error while opening database: %v", err)
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

	db, err := sql.Open("sqlite", databasePath)

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
