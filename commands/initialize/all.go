package initialize

import (
	"bctbackend/commands/common"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"
)

type initializeAllCommand struct {
	InitializeCommand
}

func NewInitializeAllCommand() *cobra.Command {
	var command *initializeAllCommand

	command = &initializeAllCommand{
		InitializeCommand: InitializeCommand{
			Command: common.Command{
				CobraCommand: &cobra.Command{
					Use:   "all",
					Short: "Initializes all components",
					Long: heredoc.Doc(`
						Initializes all components.
						Does not overwrite existing files.
					`),
					RunE: func(cmd *cobra.Command, args []string) error {
						return command.execute()
					},
					Args: cobra.NoArgs,
				},
			},
		},
	}

	return command.AsCobraCommand()
}

func (c *initializeAllCommand) execute() error {
	if err := c.generateConfigurationFile(); err != nil {
		return err
	}

	if err := c.downloadHTMLFile(false); err != nil {
		return err
	}

	if err := c.createDatabaseFile(); err != nil {
		return err
	}

	return nil
}
