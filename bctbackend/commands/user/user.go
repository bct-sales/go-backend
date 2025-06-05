package user

import (
	"github.com/spf13/cobra"
)

func NewUserCommand() *cobra.Command {
	itemCommand := cobra.Command{
		Use:   "user",
		Short: "Manage users",
		Long:  `Commands to manage users in the BCT backend system.`,
	}

	itemCommand.AddCommand(NewUserListCommand())
	itemCommand.AddCommand(NewUserAddCommand())
	itemCommand.AddCommand(NewUserShowCommand())
	itemCommand.AddCommand(NewUserSetPasswordCommand())
	itemCommand.AddCommand(NewUserRemoveCommand())
	itemCommand.AddCommand(NewUserAddSellersCommand())

	return &itemCommand
}
