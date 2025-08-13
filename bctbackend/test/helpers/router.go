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
		LogFilename: nil,
		Font: &configuration.FontConfiguration{
			Directory: os.Getenv("BCT_FONT_DIR"),
			Filename:  os.Getenv("BCT_FONT_FILE"),
			Family:    os.Getenv("BCT_FONT_FAMILY"),
		},
		BarcodeWidth:                150,
		BarcodeHeight:               30,
		GinMode:                     gin.TestMode,
		ExpiredSessionPruneInterval: 60,
	}

	server, err := server.NewServer(db, &configuration)
	if err != nil {
		panic("failed to create server")
	}

	return server
}
