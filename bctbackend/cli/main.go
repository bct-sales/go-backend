package cli

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/urfave/cli/v2"

	_ "modernc.org/sqlite"
)

const (
	DatabaseEnvironmentVariable = "BCT_DATABASE"
)

func ProcessCommandLineArguments(arguments []string) error {
	err := godotenv.Load()

	databasePath, ok := os.LookupEnv(DatabaseEnvironmentVariable)

	if !ok {
		return fmt.Errorf("environment variable %s not set", DatabaseEnvironmentVariable)
	}

	if err != nil {
		return fmt.Errorf("error while loading .env file: %v", err)
	}

	app := &cli.App{
		Commands: []*cli.Command{
			{
				Name:   "server",
				Usage:  "start REST api server",
				Action: startRestService,
			},
			{
				Name: "db",
				Subcommands: []*cli.Command{
					{
						Name:  "reset",
						Usage: "resets database; all data will be lost!",
						Action: func(context *cli.Context) error {
							return resetDatabase(databasePath)
						},
					},
					{
						Name:  "backup",
						Usage: "makes a backup",
						Action: func(context *cli.Context) error {
							arguments := context.Args()
							if arguments.Len() != 1 {
								return fmt.Errorf("expected the backup file name as argument")
							}
							targetPath := arguments.First()
							return backupDatabase(databasePath, targetPath)
						},
					},
				},
			},
		},
	}

	if err := app.Run(arguments); err != nil {
		return fmt.Errorf("error while processing command line arguments: %v", err)
	}

	return nil
}
