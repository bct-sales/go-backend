package cli

import (
	cli_barcode "bctbackend/cli/barcode"
	cli_category "bctbackend/cli/category"
	"bctbackend/cli/csv"
	cli_item "bctbackend/cli/item"
	cli_sale "bctbackend/cli/sale"
	cli_user "bctbackend/cli/user"
	"bctbackend/database/models"
	"fmt"
	"log/slog"
	"os"
	"strconv"

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
		err = fmt.Errorf("environment variable %s not set", DatabaseEnvironmentVariable)
		return err
	}

	if err != nil {
		err = fmt.Errorf("error while loading .env file: %v", err)
		return err
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

			show struct {
				id int64
			}

			remove struct {
				id int64
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
				frozen      bool
			}

			remove struct {
				id int64
			}

			show struct {
				id int64
			}
		}

		sale struct {
			add struct {
				cashierId int64
			}

			show struct {
				saleId int64
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
		Name:  "bctbackend",
		Usage: "Backend for the BCT sales site",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "verbose",
				Usage: "enable verbose output",
				Action: func(context *cli.Context, verbose bool) error {
					if verbose {
						fmt.Print("Verbose output enabled\n")
						slog.SetLogLoggerLevel(slog.LevelDebug)
					}
					return nil
				},
			},
		},
		Commands: []*cli.Command{
			{
				Name:  "server",
				Usage: "start REST api server",
				Action: func(ctx *cli.Context) error {
					return startRestService(databasePath)
				},
			},
			{
				Name:  "export",
				Usage: "export data",
				Subcommands: []*cli.Command{
					{
						Name:  "csv",
						Usage: "export data as csv",
						Subcommands: []*cli.Command{
							{
								Name:  "users",
								Usage: "export users as csv",
								Action: func(context *cli.Context) error {
									return csv.ExportUsers(databasePath)
								},
							},
							{
								Name:  "items",
								Usage: "export items as csv",
								Action: func(context *cli.Context) error {
									return csv.ExportItems(databasePath)
								},
							},
						},
					},
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
							role := options.user.add.role
							userPassword := options.user.add.password
							return cli_user.AddUser(databasePath, id, role, userPassword)
						},
					},
					{
						Name:  "remove",
						Usage: "remove a user",
						Flags: []cli.Flag{
							&cli.Int64Flag{
								Name:        "id",
								Usage:       "id of the user",
								Destination: &options.user.remove.id,
								Required:    true,
							},
						},
						Action: func(context *cli.Context) error {
							id := options.user.remove.id
							return cli_user.RemoveUser(databasePath, id)
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
						Name:  "show",
						Usage: "show information about a user",
						Flags: []cli.Flag{
							&cli.Int64Flag{
								Name:        "id",
								Usage:       "id of the user",
								Destination: &options.user.show.id,
								Required:    true,
							},
						},
						Action: func(context *cli.Context) error {
							id := options.user.show.id
							return cli_user.ShowUser(databasePath, id)
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
							&cli.BoolFlag{
								Name:        "frozen",
								Usage:       "is the item frozen?",
								Destination: &options.item.add.frozen,
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
							frozen := options.item.add.frozen

							return cli_item.AddItem(databasePath, description, price, category, seller, donation, charity, frozen)
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
					{
						Name:  "show",
						Usage: "show information about an item",
						Flags: []cli.Flag{
							&cli.Int64Flag{
								Name:        "id",
								Usage:       "id of the item",
								Destination: &options.item.show.id,
								Required:    true,
							},
						},
						Action: func(context *cli.Context) error {
							id := options.item.show.id
							return cli_item.ShowItem(databasePath, id)
						},
					},
				},
			},
			{
				Name:  "sale",
				Usage: "sale related functionality",
				Subcommands: []*cli.Command{
					{
						Name:  "list",
						Usage: "list all sales",
						Action: func(context *cli.Context) error {
							return cli_sale.ListSales(databasePath)
						},
					},
					{
						Name:  "add",
						Usage: "add a new sale",
						Flags: []cli.Flag{
							&cli.Int64Flag{
								Name:        "cashier",
								Usage:       "id of the cashier",
								Destination: &options.sale.add.cashierId,
								Required:    true,
							},
						},
						Action: func(context *cli.Context) error {
							cashierId := options.sale.add.cashierId
							items := []models.Id{}

							for i := 0; i < context.Args().Len(); i++ {
								itemId, err := strconv.ParseInt(context.Args().Get(i), 10, 64)

								if err != nil {
									err = fmt.Errorf("failed to parse item id: %v", err)
									return err
								}

								items = append(items, itemId)
							}

							return cli_sale.AddSale(databasePath, cashierId, items)
						},
					},
					{
						Name:  "show",
						Usage: "show a sale",
						Flags: []cli.Flag{
							&cli.Int64Flag{
								Name:        "sale",
								Usage:       "id of the sale",
								Destination: &options.sale.show.saleId,
								Required:    true,
							},
						},
						Action: func(context *cli.Context) error {
							saleId := options.sale.show.saleId
							return cli_sale.ShowSale(databasePath, saleId)
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
				if _, err = fmt.Fprintf(context.App.ErrWriter, "Error: %v\n", err); err != nil {
					panic(err)
				}
				os.Exit(1)
			}

			os.Exit(0)
		},
	}

	if err := app.Run(arguments); err != nil {
		err = fmt.Errorf("error while processing command line arguments: %v", err)
		return err
	}

	return nil
}
