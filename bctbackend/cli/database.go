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

func resetDatabase(context *cli.Context) error {
	db, err := sql.Open("sqlite", "../bct.db")

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

func backupDatabase(context *cli.Context) error {
	arguments := context.Args()

	if arguments.Len() != 1 {
		return fmt.Errorf("expected the backup file name as argument")
	}

	target := arguments.First()
	fmt.Printf("Backing up database to %s\n", target)

	db, err := sql.Open("sqlite", "../bct.db")

	if err != nil {
		return err
	}

	defer db.Close()

	if _, err = db.Exec("VACUUM INTO ?", target); err != nil {
		return fmt.Errorf("failed to backup database: %v", err)
	}

	fmt.Println("Database backup completed successfully")

	return nil
}
