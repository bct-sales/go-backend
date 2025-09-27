package email

import (
	"github.com/spf13/cobra"
)

func NewEmailCommand() *cobra.Command {
	command := cobra.Command{
		Use:   "email",
		Short: "Performs email related operations",
		Long:  `Commands related to emails.`,
	}

	command.AddCommand(NewTestEmailCommand())

	return &command
}
