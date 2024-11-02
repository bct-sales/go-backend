package rest

import (
	"bctbackend/database"
	"bctbackend/database/queries"
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
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

func StartRestService(databasePath string) error {
	db, err := database.ConnectToDatabase(databasePath)

	if err != nil {
		return err
	}

	defer db.Close()

	if err := enableForeignKeys(db); err != nil {
		return err
	}

	router := gin.Default()
	v1 := router.Group("/api/v1")
	v1.GET("/items", func(context *gin.Context) { getItems(context, db) })

	return router.Run("localhost:8000")
}
