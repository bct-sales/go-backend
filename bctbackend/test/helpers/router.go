//go:build test

package helpers

import (
	"bctbackend/rest"
	"database/sql"

	gin "github.com/gin-gonic/gin"
	_ "modernc.org/sqlite"
)

func CreateRestRouter() (*sql.DB, *gin.Engine) {
	db := OpenInitializedDatabase()
	gin.SetMode(gin.TestMode)
	router := gin.New()
	rest.DefineEndpoints(db, router)

	return db, router
}
