package cli

import (
	"bctbackend/database/models"
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

	var role string
	var userPassword string
	var userId models.Id

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
			{
				Name: "add",
				Subcommands: []*cli.Command{
					{
						Name:  "user",
						Usage: "add a new user",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:        "role",
								Usage:       "role of the user (admin, seller, cashier)",
								Destination: &role,
								Required:    true,
							},
							&cli.Int64Flag{
								Name:        "id",
								Usage:       "id of the user",
								Destination: &userId,
								Required:    true,
							},
							&cli.StringFlag{
								Name:        "password",
								Usage:       "password of the user",
								Destination: &userPassword,
								Required:    true,
							},
						},
						Action: func(context *cli.Context) error {
							roleId, err := models.ParseRole(role)
							if err != nil {
								return fmt.Errorf("error while parsing role: %v", err)
							}
							return addUser(databasePath, userId, roleId, userPassword)
						},
					},
				},
			},
			{
				Name: "list",
				Subcommands: []*cli.Command{
					{
						Name:  "users",
						Usage: "list all users",
						Action: func(context *cli.Context) error {
							return listUsers(databasePath)
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
