package cli

import (
	cli_item "bctbackend/cli/item"
	cli_user "bctbackend/cli/user"
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

		user struct {
			add struct {
				id       int64
				role     string
				password string
			}

			setPassword struct {
				id       int64
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
				Name:  "user",
				Usage: "user related functionality",
				Subcommands: []*cli.Command{
					{
						Name:  "add",
						Usage: "add a new user",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:        "role",
								Usage:       "role of the user (admin, seller, cashier)",
								Destination: &options.user.add.role,
								Required:    true,
							},
							&cli.Int64Flag{
								Name:        "id",
								Usage:       "id of the user",
								Destination: &options.user.add.id,
								Required:    true,
							},
							&cli.StringFlag{
								Name:        "password",
								Usage:       "password of the user",
								Destination: &options.user.add.password,
								Required:    true,
							},
						},
						Action: func(context *cli.Context) error {
							id := options.user.add.id
							roleId, err := models.ParseRole(options.user.add.role)
							userPassword := options.user.add.password
							if err != nil {
								return fmt.Errorf("error while parsing role: %v", err)
							}
							return cli_user.AddUser(databasePath, id, roleId, userPassword)
						},
					},
					{
						Name:  "list",
						Usage: "list all users",
						Action: func(context *cli.Context) error {
							return cli_user.ListUsers(databasePath)
						},
					},
					{
						Name:  "set-password",
						Usage: "set password for a user",
						Flags: []cli.Flag{
							&cli.Int64Flag{
								Name:        "id",
								Usage:       "id of the user",
								Destination: &options.user.setPassword.id,
								Required:    true,
							},
							&cli.StringFlag{
								Name:        "password",
								Usage:       "new password for the user",
								Destination: &options.user.setPassword.password,
								Required:    true,
							},
						},
						Action: func(context *cli.Context) error {
							id := options.user.setPassword.id
							password := options.user.setPassword.password
							return cli_user.SetPassword(databasePath, id, password)
						},
					},
				},
			},
			{
				Name:  "item",
				Usage: "item related functionality",
				Subcommands: []*cli.Command{
					{
						Name:  "list",
						Usage: "list all items",
						Action: func(context *cli.Context) error {
							return cli_item.ListItems(databasePath)
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
