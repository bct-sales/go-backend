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

	command.AddCommand(NewItemListCommand())
	command.AddCommand(NewItemAddCommand())
	command.AddCommand(NewItemShowCommand())
	command.AddCommand(NewItemFreezeCommand())
	command.AddCommand(NewItemHideCommand())
	command.AddCommand(NewItemUnfreezeCommand())
	command.AddCommand(NewItemUnhideCommand())
	command.AddCommand(NewItemRemoveCommand())
	command.AddCommand(NewItemCopyCommand())

	return &command
}
