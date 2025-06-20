package initialize

import (
	"bctbackend/commands/common"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
			},
		},
	}

	return command.AsCobraCommand()
}

func (c *InitializeCommand) execute() error {
	if err := viper.SafeWriteConfig(); err != nil {
		c.Printf("Failed to create configuration file: %v\n", err)
		return err
	}

	c.Printf("Configuration file created successfully at %s\n", viper.ConfigFileUsed())
	return nil
}
