package commands

import (
	"bctbackend/commands/common"
	"bctbackend/version"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"
)

type versionCommand struct {
	common.Command
}

func NewVersionCommand() *cobra.Command {
	var command *versionCommand

	command = &versionCommand{
		Command: common.Command{
			CobraCommand: &cobra.Command{
				Use:   "version",
				Short: "Shows version",
				Long: heredoc.Doc(`
						Shows version.
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

func (c *versionCommand) execute() error {
	c.Printf("%s\n", version.GetVersionString())

	return nil
}
