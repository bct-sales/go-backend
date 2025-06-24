//go:build test

package helpers

import (
	"bctbackend/server"
	"database/sql"
	"os"

	gin "github.com/gin-gonic/gin"
)

func CreateRestRouter(db *sql.DB) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	configuration := server.Configuration{
		FontDirectory: os.Getenv("BCT_FONT_DIR"),
		FontFilename:  os.Getenv("BCT_FONT_FILE"),
		FontFamily:    os.Getenv("BCT_FONT_FAMILY"),
		BarcodeWidth:  150,
		BarcodeHeight: 30,
	}
	server.DefineEndpoints(db, router, &configuration)

	return router
}
