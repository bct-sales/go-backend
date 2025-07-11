package sale

import (
	"github.com/spf13/cobra"
)

func NewSaleCommand() *cobra.Command {
	command := cobra.Command{
		Use:   "sale",
		Short: "Manage sales",
		Long:  `Commands to manage sales in the BCT backend system.`,
	}

	command.AddCommand(NewSaleListCommand())
	command.AddCommand(NewSaleAddCommand())
	command.AddCommand(NewSaleShowCommand())
	command.AddCommand(NewRemoveAllSalesCommand())

	return &command
}
