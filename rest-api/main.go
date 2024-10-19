package main

import (
	queries "bctrest/db/queries"
	"fmt"
	"net/http"

	"database/sql"

	"github.com/gin-gonic/gin"

	_ "modernc.org/sqlite"
)

func getItems(context *gin.Context, db *sql.DB) {
	items, err := queries.GetItems(db)

	if err != nil {
		context.AbortWithStatus(http.StatusInternalServerError)
	}

	context.IndentedJSON(http.StatusOK, items)
}

func startRestService() {
	db, err := sql.Open("sqlite", "../bct.db")

	if err != nil {
		fmt.Println(err)
		return
	}

	defer db.Close()

	router := gin.Default()
	v1 := router.Group("/api/v1")
	v1.GET("/items", func(context *gin.Context) { getItems(context, db) })

	router.Run("localhost:8000")
}

func main() {
	startRestService()
}
