package item

import (
	"github.com/spf13/cobra"
)

func NewItemCommand() *cobra.Command {
	itemCommand := cobra.Command{
		Use:   "item",
		Short: "Manage items",
		Long:  `Commands to manage items in the BCT backend system.`,
	}

	itemCommand.AddCommand(NewItemListCommand())

	return &itemCommand
}
