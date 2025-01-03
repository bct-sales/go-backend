package cli

import (
	cli_barcode "bctbackend/cli/barcode"
	cli_category "bctbackend/cli/category"
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

		item struct {
			add struct {
				description string
				category    int64
				price       int64
				seller      int64
				donation    bool
				charity     bool
			}

			remove struct {
				id int64
			}
		}

		barcode struct {
			raw struct {
				data       string
				outputPath string
				width      int
				height     int
			}
		}
	}

	app := &cli.App{
		Commands: []*cli.Command{
			{
				Name:  "server",
				Usage: "start REST api server",
				Action: func(ctx *cli.Context) error {
					return startRestService(databasePath)
				},
			},
			{
				Name:  "db",
				Usage: "database related functionality",
				Subcommands: []*cli.Command{
					{
						Name:  "reset",
						Usage: "resets database; all data will be lost!",
						Action: func(context *cli.Context) error {
							return resetDatabase(databasePath)
						},
					},
					{
						Name:  "dummy",
						Usage: "resets database and populates it with dummy data; all data will be lost!",
						Action: func(context *cli.Context) error {
							return resetDatabaseAndFillWithDummyData(databasePath)
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
					{
						Name:  "add",
						Usage: "add a new item",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:        "description",
								Usage:       "description of the item",
								Destination: &options.item.add.description,
								Required:    true,
							},
							&cli.Int64Flag{
								Name:        "category",
								Usage:       "category of the item",
								Destination: &options.item.add.category,
								Required:    true,
							},
							&cli.Int64Flag{
								Name:        "price-in-cents",
								Usage:       "price of the item in cents",
								Destination: &options.item.add.price,
								Required:    true,
							},
							&cli.Int64Flag{
								Name:        "seller",
								Usage:       "id of the seller",
								Destination: &options.item.add.seller,
								Required:    true,
							},
							&cli.BoolFlag{
								Name:        "donation",
								Usage:       "is the item a donation?",
								Destination: &options.item.add.donation,
								Value:       false,
							},
							&cli.BoolFlag{
								Name:        "charity",
								Usage:       "is the item a charity?",
								Destination: &options.item.add.charity,
								Value:       false,
							},
						},
						Action: func(context *cli.Context) error {
							description := options.item.add.description
							category := options.item.add.category
							price := options.item.add.price
							seller := options.item.add.seller
							donation := options.item.add.donation
							charity := options.item.add.charity

							return cli_item.AddItem(databasePath, description, price, category, seller, donation, charity)
						},
					},
					{
						Name:  "remove",
						Usage: "remove an item",
						Flags: []cli.Flag{
							&cli.Int64Flag{
								Name:        "id",
								Usage:       "id of the item",
								Destination: &options.item.remove.id,
								Required:    true,
							},
						},
						Action: func(context *cli.Context) error {
							id := options.item.remove.id
							return cli_item.RemoveItem(databasePath, id)
						},
					},
				},
			},
			{
				Name:  "category",
				Usage: "category related functionality",
				Subcommands: []*cli.Command{
					{
						Name:  "list",
						Usage: "list all categories",
						Action: func(context *cli.Context) error {
							return cli_category.ListCategories(databasePath)
						},
					},
					{
						Name:  "counts",
						Usage: "list the number of items in each category",
						Action: func(context *cli.Context) error {
							return cli_category.ListCategoryCounts(databasePath)
						},
					},
				},
			},
			{
				Name:  "barcode",
				Usage: "barcode related functionality",
				Subcommands: []*cli.Command{
					{
						Name:  "raw",
						Usage: "generate a raw barcode",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:        "data",
								Usage:       "data to encode in the barcode",
								Destination: &options.barcode.raw.data,
								Required:    true,
							},
							&cli.StringFlag{
								Name:        "output",
								Usage:       "filename to save the barcode to",
								Destination: &options.barcode.raw.outputPath,
								Required:    true,
							},
							&cli.IntFlag{
								Name:        "width",
								Usage:       "width of the barcode",
								Destination: &options.barcode.raw.width,
								Value:       200,
							},
							&cli.IntFlag{
								Name:        "height",
								Usage:       "height of the barcode",
								Destination: &options.barcode.raw.height,
								Value:       100,
							},
						},
						Action: func(context *cli.Context) error {
							return cli_barcode.GenerateRawBarcode(
								options.barcode.raw.data,
								options.barcode.raw.outputPath,
								options.barcode.raw.width,
								options.barcode.raw.height,
							)
						},
					},
					{
						Name:  "pdf",
						Usage: "generate pdf with barcodes",
						Action: func(context *cli.Context) error {
							return cli_barcode.GeneratePdf()
						},
					},
				},
			},
		},
		ExitErrHandler: func(context *cli.Context, err error) {
			if err != nil {
				fmt.Fprintf(context.App.ErrWriter, "Error: %v\n", err)
				os.Exit(1)
			}

			os.Exit(0)
		},
	}

	if err := app.Run(arguments); err != nil {
		return fmt.Errorf("error while processing command line arguments: %v", err)
	}

	return nil
}
