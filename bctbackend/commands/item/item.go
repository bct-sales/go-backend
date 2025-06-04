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
	itemCommand.AddCommand(NewItemAddCommand())
	itemCommand.AddCommand(NewItemShowCommand())
	itemCommand.AddCommand(NewItemFreezeCommand())
	itemCommand.AddCommand(NewItemHideCommand())

	return &itemCommand
}
