package commands

import (
	"bctbackend/commands/common"
	"bctbackend/ui"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"
)

type UserInterfaceCommand struct {
	common.Command
}

func NewUserInterfaceCommand() *cobra.Command {
	var command *UserInterfaceCommand

	command = &UserInterfaceCommand{
		Command: common.Command{
			CobraCommand: &cobra.Command{
				Use:   "ui",
				Short: "Start UI [work in progress]",
				Long: heredoc.Doc(`
						Starts UI. Not finished, at all.
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

func (c *UserInterfaceCommand) execute() error {
	return c.WithOpenedDatabase(ui.Start)
}
