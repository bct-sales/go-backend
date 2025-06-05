package cli

import (
	cli_barcode "bctbackend/cli/barcode"
	cli_category "bctbackend/cli/category"
	"bctbackend/cli/csv"
	. "bctbackend/cli/database"
	. "bctbackend/cli/sale"
	. "bctbackend/cli/user"
	config "bctbackend/configuration"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/urfave/cli/v2"
)

const (
	DatabaseEnvironmentVariable = "BCT_DATABASE"
)

var ErrInvalidZoneFormat = fmt.Errorf("invalid zone format")

func parseZones(zoneStrings []string) ([]int, error) {
	result := []int{}

	for _, zoneString := range zoneStrings {
		parts := strings.Split(zoneString, ",")

		for _, part := range parts {
			endpoints := strings.Split(part, "-")
			if len(endpoints) == 1 {
				zone, err := strconv.Atoi(strings.TrimSpace(endpoints[0]))
				if err != nil {
					return nil, fmt.Errorf("invalid zone format %s: %w, %w", part, err, ErrInvalidZoneFormat)
				}
				result = append(result, zone)
			} else if len(endpoints) == 2 {
				start, err := strconv.Atoi(strings.TrimSpace(endpoints[0]))
				if err != nil {
					return nil, fmt.Errorf("invalid zone format %s: %w, %w", part, err, ErrInvalidZoneFormat)
				}
				end, err := strconv.Atoi(strings.TrimSpace(endpoints[1]))
				if err != nil {
					return nil, fmt.Errorf("invalid zone format %s: %w, %w", part, err, ErrInvalidZoneFormat)
				}
				if start >= end {
					return nil, fmt.Errorf("invalid zone format %s: %w, %w", part, err, ErrInvalidZoneFormat)
				}
				for i := start; i <= end; i++ {
					result = append(result, i)
				}
			} else {
				return nil, fmt.Errorf("invalid zone format %s: %w", part, ErrInvalidZoneFormat)
			}
		}
	}

	slices.Sort(result)
	result = slices.Compact(result)

	return result, nil
}

func ProcessCommandLineArguments(arguments []string) error {
	if err := godotenv.Load(); err != nil {
		return fmt.Errorf("error while loading .env file: %w", err)
	}

	var options struct {
		configurationPath string

		global *config.Configuration

		db struct {
			backup struct {
				target string
			}

			init struct {
				noCategories bool
			}

			reset struct {
				noCategories bool
			}
		}

		export struct {
			csv struct {
				items struct {
					showHidden bool
				}
			}
		}

		user struct {
			add struct {
				id       int64
				role     string
				password string
			}

			addSellers struct {
				seed           uint64
				zones          []int
				sellersPerZone int
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
			}

			list struct {
				showHidden bool
			}

			remove struct {
				id int64
			}

			show struct {
				id int64
			}

			freeze struct {
				id int64
			}

			unfreeze struct {
				id int64
			}

			hide struct {
				id int64
			}

			unhide struct {
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

		category struct {
			add struct {
				id   int64
				name string
			}

			counts struct {
				includeHiddenItems bool
			}
		}
	}

	//exhaustruct:ignore
	app := &cli.App{
		Name:  "bctbackend",
		Usage: "Backend for the BCT sales site",
		Flags: []cli.Flag{
			//exhaustruct:ignore
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
			//exhaustruct:ignore
			&cli.StringFlag{
				Name:        "config",
				Usage:       "path to the configuration file",
				Value:       "config.yaml",
				Destination: &options.configurationPath,
			},
		},
		Before: func(context *cli.Context) error {
			absoluteConfigurationPath, err := filepath.Abs(options.configurationPath)
			if err != nil {
				return cli.Exit("Failed to get absolute path for configuration file: "+err.Error(), 1)
			}

			configuration, err := config.LoadConfigurationFromFile(absoluteConfigurationPath)
			if err != nil {
				return cli.Exit(
					fmt.Sprintf("Failed to load configuration file %s; error: %s",
						absoluteConfigurationPath,
						err.Error()),
					1)
			}

			options.global = configuration

			return nil
		},
		Commands: []*cli.Command{
			{
				Name:  "server",
				Usage: "start REST api server",
				Action: func(ctx *cli.Context) error {
					return startRestService(options.global.DatabasePath)
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
									return csv.ExportUsers(options.global.DatabasePath)
								},
							},
							{
								Name:  "items",
								Usage: "export items as csv",
								Flags: []cli.Flag{
									//exhaustruct:ignore
									&cli.BoolFlag{
										Name:        "show-hidden",
										Usage:       "show hidden items",
										Destination: &options.export.csv.items.showHidden,
										Value:       false,
									},
								},
								Action: func(context *cli.Context) error {
									showHidden := options.export.csv.items.showHidden
									return csv.ExportItems(options.global.DatabasePath, showHidden)
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
						Name:  "init",
						Usage: "creates new database; refuses to overwrite existing database",
						Flags: []cli.Flag{
							//exhaustruct:ignore
							&cli.BoolFlag{
								Name:        "no-categories",
								Usage:       "do not add default categories",
								Destination: &options.db.init.noCategories,
								Value:       false,
							},
						},
						Action: func(context *cli.Context) error {
							addCategories := !options.db.init.noCategories

							return InitializeDatabase(options.global.DatabasePath, addCategories)
						},
					},
					{
						Name:  "reset",
						Usage: "resets database; all data will be lost!",
						Flags: []cli.Flag{
							//exhaustruct:ignore
							&cli.BoolFlag{
								Name:        "no-categories",
								Usage:       "do not add default categories",
								Destination: &options.db.reset.noCategories,
								Value:       false,
							},
						},
						Action: func(context *cli.Context) error {
							addCategories := !options.db.reset.noCategories

							return ResetDatabase(options.global.DatabasePath, addCategories)
						},
					},
					{
						Name:  "dummy",
						Usage: "resets database and populates it with dummy data; all data will be lost!",
						Action: func(context *cli.Context) error {
							return ResetDatabaseAndFillWithDummyData(options.global.DatabasePath)
						},
					},
					{
						Name:  "backup",
						Usage: "makes a backup",
						Flags: []cli.Flag{
							//exhaustruct:ignore
							&cli.StringFlag{
								Name:        "target",
								Usage:       "filename of the backup",
								Destination: &options.db.backup.target,
								Required:    true,
							},
						},
						Action: func(context *cli.Context) error {
							targetPath := options.db.backup.target
							return BackupDatabase(options.global.DatabasePath, targetPath)
						},
					},
				},
			},
			{
				Name:  "user",
				Usage: "user related functionality",
				Subcommands: []*cli.Command{
					{
						Name:  "add-sellers",
						Usage: "add sellers with random passwords",
						Flags: []cli.Flag{
							//exhaustruct:ignore
							&cli.Uint64Flag{
								Name:        "seed",
								Usage:       "seed for the random number generator",
								Destination: &options.user.addSellers.seed,
								Required:    false,
							},
							//exhaustruct:ignore
							&cli.StringSliceFlag{
								Name:     "zones",
								Usage:    "zones for the sellers",
								Required: true,
							},
							//exhaustruct:ignore
							&cli.IntFlag{
								Name:        "per-zone",
								Usage:       "number of sellers per zone",
								Destination: &options.user.addSellers.sellersPerZone,
								Required:    true,
							},
						},
						Action: func(context *cli.Context) error {
							seed := options.user.addSellers.seed
							if seed == 0 {
								seed = uint64(time.Now().UnixNano())
							}

							zones, err := parseZones(context.StringSlice("zones"))
							if err != nil {
								return err
							}
							sellersPerZone := options.user.addSellers.sellersPerZone

							return AddSellers(options.global.DatabasePath, seed, zones, sellersPerZone)
						},
					},
					{
						Name:  "remove",
						Usage: "remove a user",
						Flags: []cli.Flag{
							//exhaustruct:ignore
							&cli.Int64Flag{
								Name:        "id",
								Usage:       "id of the user",
								Destination: &options.user.remove.id,
								Required:    true,
							},
						},
						Action: func(context *cli.Context) error {
							id := models.Id(options.user.remove.id)
							return RemoveUser(options.global.DatabasePath, id)
						},
					},
					{
						Name:  "list",
						Usage: "list all users",
						Action: func(context *cli.Context) error {
							return ListUsers(options.global.DatabasePath)
						},
					},
					{
						Name:  "show",
						Usage: "show information about a user",
						Flags: []cli.Flag{
							//exhaustruct:ignore
							&cli.Int64Flag{
								Name:        "id",
								Usage:       "id of the user",
								Destination: &options.user.show.id,
								Required:    true,
							},
						},
						Action: func(context *cli.Context) error {
							id := models.Id(options.user.show.id)
							return ShowUser(options.global.DatabasePath, id)
						},
					},
					{
						Name:  "set-password",
						Usage: "set password for a user",
						Flags: []cli.Flag{
							//exhaustruct:ignore
							&cli.Int64Flag{
								Name:        "id",
								Usage:       "id of the user",
								Destination: &options.user.setPassword.id,
								Required:    true,
							},
							//exhaustruct:ignore
							&cli.StringFlag{
								Name:        "password",
								Usage:       "new password for the user",
								Destination: &options.user.setPassword.password,
								Required:    true,
							},
						},
						Action: func(context *cli.Context) error {
							id := models.Id(options.user.setPassword.id)
							password := options.user.setPassword.password
							return SetPassword(options.global.DatabasePath, id, password)
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
							return ListSales(options.global.DatabasePath)
						},
					},
					{
						Name:  "add",
						Usage: "add a new sale",
						Flags: []cli.Flag{
							//exhaustruct:ignore
							&cli.Int64Flag{
								Name:        "cashier",
								Usage:       "id of the cashier",
								Destination: &options.sale.add.cashierId,
								Required:    true,
							},
						},
						Action: func(context *cli.Context) error {
							cashierId := models.Id(options.sale.add.cashierId)
							items := []models.Id{}

							for i := range context.Args().Len() {
								itemId, err := models.ParseId(context.Args().Get(i))

								if err != nil {
									return fmt.Errorf("failed to parse item id: %w", err)
								}

								items = append(items, itemId)
							}

							return AddSale(options.global.DatabasePath, cashierId, items)
						},
					},
					{
						Name:  "show",
						Usage: "show a sale",
						Flags: []cli.Flag{
							//exhaustruct:ignore
							&cli.Int64Flag{
								Name:        "sale",
								Usage:       "id of the sale",
								Destination: &options.sale.show.saleId,
								Required:    true,
							},
						},
						Action: func(context *cli.Context) error {
							saleId := models.Id(options.sale.show.saleId)
							return ShowSale(options.global.DatabasePath, saleId)
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
							return cli_category.ListCategories(options.global.DatabasePath)
						},
					},
					{
						Name:  "counts",
						Usage: "list the number of items in each category",
						Flags: []cli.Flag{
							//exhaustruct:ignore
							&cli.BoolFlag{
								Name:        "include-hidden",
								Usage:       "include hidden items",
								Destination: &options.category.counts.includeHiddenItems,
								Value:       false,
							},
						},
						Action: func(context *cli.Context) error {
							itemSelection := queries.ItemSelectionFromBool(options.category.counts.includeHiddenItems)

							return cli_category.ListCategoryCounts(options.global.DatabasePath, itemSelection)
						},
					},
					{
						Name:  "add",
						Usage: "add a new category",
						Flags: []cli.Flag{
							//exhaustruct:ignore
							&cli.Int64Flag{
								Name:        "id",
								Usage:       "id of the category",
								Destination: &options.category.add.id,
								Required:    true,
							},
							//exhaustruct:ignore
							&cli.StringFlag{
								Name:        "name",
								Usage:       "name of the category",
								Destination: &options.category.add.name,
								Required:    true,
							},
						},
						Action: func(context *cli.Context) error {
							id := models.Id(options.category.add.id)
							name := options.category.add.name

							return cli_category.AddCategory(options.global.DatabasePath, id, name)
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
							//exhaustruct:ignore
							&cli.StringFlag{
								Name:        "data",
								Usage:       "data to encode in the barcode",
								Destination: &options.barcode.raw.data,
								Required:    true,
							},
							//exhaustruct:ignore
							&cli.StringFlag{
								Name:        "output",
								Usage:       "filename to save the barcode to",
								Destination: &options.barcode.raw.outputPath,
								Required:    true,
							},
							//exhaustruct:ignore
							&cli.IntFlag{
								Name:        "width",
								Usage:       "width of the barcode",
								Destination: &options.barcode.raw.width,
								Value:       200,
							},
							//exhaustruct:ignore
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
	}

	if err := app.Run(arguments); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	return nil
}
