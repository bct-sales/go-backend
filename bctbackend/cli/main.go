package cli

import (
	config "bctbackend/configuration"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

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
		},
	}

	if err := app.Run(arguments); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	return nil
}
