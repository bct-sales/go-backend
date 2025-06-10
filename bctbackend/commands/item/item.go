package item

import (
	"github.com/spf13/cobra"
)

func NewItemCommand() *cobra.Command {
	command := cobra.Command{
		Use:   "item",
		Short: "Manage items",
		Long:  `Commands to manage items in the BCT backend system.`,
	}

	command.AddCommand(NewListItemsCommand())
	command.AddCommand(NewAddItemCommand())
	command.AddCommand(NewShowItemCommand())
	command.AddCommand(NewFreezeItemCommand())
	command.AddCommand(NewHideItemCommand())
	command.AddCommand(NewUnfreezeItemCommand())
	command.AddCommand(NewUnhideItemCommand())
	command.AddCommand(NewRemoveItemCommand())
	command.AddCommand(NewCopyItemCommand())

	return &command
}
