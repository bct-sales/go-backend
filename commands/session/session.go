package session

import (
	"github.com/spf13/cobra"
)

func NewSessionCommand() *cobra.Command {
	command := cobra.Command{
		Use:   "session",
		Short: "Manage sessions",
		Long:  `Commands to manage sessions in the BCT backend system.`,
	}

	command.AddCommand(NewPruneSessionsCommand())
	command.AddCommand(NewListSessionsCommand())

	return &command
}
