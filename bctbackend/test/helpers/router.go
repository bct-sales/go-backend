//go:build test

package helpers

import (
	"bctbackend/rest"
	"database/sql"
	"os"

	gin "github.com/gin-gonic/gin"
)

func CreateRestRouter(db *sql.DB) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	configuration := rest.Configuration{
		FontDirectory: os.Getenv("BCT_FONT_DIR"),
		FontFilename:  os.Getenv("BCT_FONT_FILE"),
		FontFamily:    os.Getenv("BCT_FONT_FAMILY"),
		BarcodeWidth:  150,
		BarcodeHeight: 30,
	}
	rest.DefineEndpoints(db, router, &configuration)

	return router
}
