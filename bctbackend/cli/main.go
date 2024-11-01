package cli

import (
	"fmt"

	"github.com/urfave/cli/v2"

	_ "modernc.org/sqlite"
)

func ProcessCommandLineArguments(arguments []string) error {
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
							return resetDatabase()
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
							target := arguments.First()
							return backupDatabase(target)
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
