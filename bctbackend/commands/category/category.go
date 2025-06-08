package category

import (
	"github.com/spf13/cobra"
)

func NewCategoryCommand() *cobra.Command {
	command := cobra.Command{
		Use:   "category",
		Short: "Manage categories",
		Long:  `Commands to manage categories in the BCT backend system.`,
	}

	command.AddCommand(NewCategoryListCommand())
	command.AddCommand(NewCategoryCountCommand())

	return &command
}
