//go:build test

package helpers

import (
	"bctbackend/clock"
	"bctbackend/server"
	"bctbackend/server/configuration"
	"database/sql"
	"os"

	gin "github.com/gin-gonic/gin"
)

func CreateRestServer(db *sql.DB, clock *clock.ManualClock) *server.Server {
	configuration := configuration.Configuration{
		Log: nil,
		LabelGeneration: &configuration.LabelGenerationConfiguration{
			BarcodeWidth:  150,
			BarcodeHeight: 30,
			Font: &configuration.FontConfiguration{
				Directory: os.Getenv("BCT_FONT_DIR"),
				Filename:  os.Getenv("BCT_FONT_FILE"),
				Family:    os.Getenv("BCT_FONT_FAMILY"),
			},
		},
		GinMode:                     gin.TestMode,
		ExpiredSessionPruneInterval: 60,
	}

	server, err := server.NewServer(clock, db, &configuration)
	if err != nil {
		panic("failed to create server")
	}

	return server
}
