//go:build test

package helpers

import (
	"bctbackend/server"
	"bctbackend/server/configuration"
	"database/sql"
	"os"

	gin "github.com/gin-gonic/gin"
)

func CreateRestServer(db *sql.DB) *server.Server {
	configuration := configuration.Configuration{
		FontDirectory: os.Getenv("BCT_FONT_DIR"),
		FontFilename:  os.Getenv("BCT_FONT_FILE"),
		FontFamily:    os.Getenv("BCT_FONT_FAMILY"),
		BarcodeWidth:  150,
		BarcodeHeight: 30,
		GinMode:       gin.TestMode,
	}

	server := server.NewServer(db, &configuration)

	return server
}
