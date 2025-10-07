//go:build test

package helpers

import (
	"bctbackend/clock"
	"bctbackend/server"
	"bctbackend/server/configuration"
	"database/sql"

	gin "github.com/gin-gonic/gin"
)

func CreateRestServer(db *sql.DB, clock *clock.ManualClock) *server.Server {
	configuration := configuration.Configuration{
		Log: nil,
		LabelGeneration: &configuration.LabelGenerationConfiguration{
			BarcodeWidth:  150,
			BarcodeHeight: 30,
		},
		Server: &configuration.ServerConfiguration{
			GinMode:                     gin.TestMode,
			ExpiredSessionPruneInterval: 60,
		},
	}

	server, err := server.NewServer(clock, db, &configuration)
	if err != nil {
		panic("failed to create server")
	}

	return server
}
