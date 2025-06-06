package sale

import (
	"bctbackend/cli/formatting"
	"bctbackend/commands/common"
	"bctbackend/database/models"
	"database/sql"
	"fmt"

	"github.com/spf13/cobra"
)

func NewSaleShowCommand() *cobra.Command {
	command := cobra.Command{
		Use:   "show",
		Short: "Show a sale",
		Long:  `This command shows the details of a specific sale by its ID.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return common.WithOpenedDatabase(cmd.ErrOrStderr(), func(db *sql.DB) error {
				saleId, err := models.ParseId(args[0])
				if err != nil {
					fmt.Fprintf(cmd.ErrOrStderr(), "Invalid sale ID: %v\n", err)
					return fmt.Errorf("invalid sale ID: %w", err)
				}

				err = formatting.PrintSale(db, saleId)
				if err != nil {
					fmt.Fprintf(cmd.ErrOrStderr(), "Failed to print sale: %v\n", err)
					return err
				}

				return nil
			})
		},
	}

	return &command
}
