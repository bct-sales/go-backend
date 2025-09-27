package initialize

import (
	"bctbackend/commands/common"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"
)

type initializeHtmlCommand struct {
	InitializeCommand
}

func NewInitializeHtmlCommand() *cobra.Command {
	var command *initializeHtmlCommand

	command = &initializeHtmlCommand{
		InitializeCommand: InitializeCommand{
			Command: common.Command{
				CobraCommand: &cobra.Command{
					Use:   "html",
					Short: "Download html",
					Long: heredoc.Doc(`
						Downloads index.html from GitHub.
						Overwrites existing index.html.
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

func (c *initializeHtmlCommand) execute() error {
	return c.downloadHTMLFile(true)
}
