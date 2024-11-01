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
