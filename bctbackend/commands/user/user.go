package user

import (
	"github.com/spf13/cobra"
)

func NewUserCommand() *cobra.Command {
	command := cobra.Command{
		Use:   "user",
		Short: "Manage users",
		Long:  `Commands to manage users in the BCT backend system.`,
	}

	command.AddCommand(NewUserListCommand())
	command.AddCommand(NewUserAddCommand())
	command.AddCommand(NewUserShowCommand())
	command.AddCommand(NewUserSetPasswordCommand())
	command.AddCommand(NewUserRemoveCommand())
	command.AddCommand(NewUserAddSellersCommand())

	return &command
}
