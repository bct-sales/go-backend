package commands

import (
	"bctbackend/commands/common"
	"bctbackend/ui"
	"context"

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
					return command.execute(cmd.Context())
				},
				Args: cobra.NoArgs,
			},
		},
	}

	return command.AsCobraCommand()
}

func (c *UserInterfaceCommand) execute(ctx context.Context) error {
	return c.WithOpenedDatabase(ctx, ui.Start)
}
