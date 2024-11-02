package cli

import (
	cli_add "bctbackend/cli/add"
	cli_list "bctbackend/cli/list"
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

	var options struct {
		db struct {
			backup struct {
				target string
			}
		}

		add struct {
			user struct {
				id       int64
				role     string
				password string
			}
		}
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
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:        "target",
								Usage:       "filename of the backup",
								Destination: &options.db.backup.target,
								Required:    true,
							},
						},
						Action: func(context *cli.Context) error {
							targetPath := options.db.backup.target
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
								Destination: &options.add.user.role,
								Required:    true,
							},
							&cli.Int64Flag{
								Name:        "id",
								Usage:       "id of the user",
								Destination: &options.add.user.id,
								Required:    true,
							},
							&cli.StringFlag{
								Name:        "password",
								Usage:       "password of the user",
								Destination: &options.add.user.password,
								Required:    true,
							},
						},
						Action: func(context *cli.Context) error {
							id := options.add.user.id
							roleId, err := models.ParseRole(options.add.user.role)
							userPassword := options.add.user.password
							if err != nil {
								return fmt.Errorf("error while parsing role: %v", err)
							}
							return cli_add.AddUser(databasePath, id, roleId, userPassword)
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
							return cli_list.ListUsers(databasePath)
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
