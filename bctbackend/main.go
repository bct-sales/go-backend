package main

import (
	database "bctbackend/database"
	queries "bctbackend/database/queries"
	"fmt"
	"log"
	"net/http"
	"os"

	"database/sql"

	"github.com/gin-gonic/gin"
	"github.com/urfave/cli/v2"

	_ "modernc.org/sqlite"
)

func getItems(context *gin.Context, db *sql.DB) {
	items, err := queries.GetItems(db)

	if err != nil {
		context.AbortWithStatus(http.StatusInternalServerError)
	}

	context.IndentedJSON(http.StatusOK, items)
}

func enableForeignKeys(db *sql.DB) error {
	_, err := db.Exec("PRAGMA foreign_keys = ON")
	return err
}

func startRestService() {
	db, err := sql.Open("sqlite", "../bct.db")

	if err != nil {
		fmt.Println(err)
		return
	}

	defer db.Close()

	if err := enableForeignKeys(db); err != nil {
		fmt.Println(err)
		return
	}

	router := gin.Default()
	v1 := router.Group("/api/v1")
	v1.GET("/items", func(context *gin.Context) { getItems(context, db) })

	router.Run("localhost:8000")
}

func startRestServiceHandler(context *cli.Context) error {
	startRestService()
	return nil
}

func resetDatabaseHandler(context *cli.Context) error {
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

func main() {
	app := &cli.App{
		Commands: []*cli.Command{
			{
				Name:   "server",
				Usage:  "start REST api server",
				Action: startRestServiceHandler,
			},
			{
				Name: "db",
				Subcommands: []*cli.Command{
					{
						Name:   "reset",
						Usage:  "resets database; all data will be lost!",
						Action: resetDatabaseHandler,
					},
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatalf("Error while parsing command line arguments: %v", err)
	}
}
