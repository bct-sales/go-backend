//go:build test

package helpers

import (
	"bctbackend/clock"
	"bctbackend/logging"
	"bctbackend/server"
	"bctbackend/server/configuration"
	"database/sql"
	"log/slog"
	"os"

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

	slogger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	logger := logging.NewSloggerWrapper(slogger)
	server, err := server.NewServer(clock, db, logger, &configuration)
	if err != nil {
		panic("failed to create server")
	}

	return server
}
