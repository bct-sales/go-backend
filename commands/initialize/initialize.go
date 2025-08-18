package initialize

import (
	"bctbackend/commands/common"
	"os"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"
)

type InitializeCommand struct {
	common.Command
}

func NewInitializeCommand() *cobra.Command {
	var command *InitializeCommand

	command = &InitializeCommand{
		Command: common.Command{
			CobraCommand: &cobra.Command{
				Use:   "init",
				Short: "Creates configuration file",
				Long: heredoc.Doc(`
							This command creates a configuration file for the BCT application.
							It will create a file named 'bctconfig.yaml' in the current directory.
					   `),
				RunE: func(cmd *cobra.Command, args []string) error {
					return command.execute()
				},
				Args: cobra.NoArgs,
			},
		},
	}

	return command.AsCobraCommand()
}

func (command *InitializeCommand) execute() error {
	contents := heredoc.Doc(`
		database: "bct.db"
		labelGeneration:
		  barcode:
		    width: 150
		    height: 25
		  font:
		    directory: "."
		    filename: "noto.ttf"
		    family: "Noto"
		server:
		  port: 80
		  html: index.html
		  pruneExpiredSessionsInterval: 3600
		debug: true
		log:
		  file: "bct.log"
		  maxSizeMegabytes: 10
		  maxBackups: 3
		  maxAgeDays: 28
		  compression: false
	`)

	err := os.WriteFile("bctconfig.yaml", []byte(contents), 0644)
	if err != nil {
		command.PrintErrorf("Failed to create configuration file: %v\n", err)
		return err
	}

	command.Printf("Configuration file created successfully\n")
	return nil
}
