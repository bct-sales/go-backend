package initialize

import (
	"bctbackend/algorithms"
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

func (c *InitializeCommand) execute() error {
	if err := c.generateConfigurationFile(); err != nil {
		return err
	}
	return nil
}

func (c *InitializeCommand) generateConfigurationFile() error {
	fileExists, err := algorithms.FileExists("bctconfig.yaml")
	if err != nil {
		c.PrintErrorf("Failed to check if configuration file exists: %v\n", err)
		return err
	}
	if fileExists {
		c.Printf("Configuration file already exists; I will not overwrite it\n")
		return nil
	}

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

	if err := os.WriteFile("bctconfig.yaml", []byte(contents), 0644); err != nil {
		c.PrintErrorf("Failed to create configuration file: %v\n", err)
		return err
	}

	c.Printf("Configuration file created successfully\n")
	return nil
}
