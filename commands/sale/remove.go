package sale

import (
	"bctbackend/commands/common"
	"bctbackend/database/queries"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"
)

type removeSalesCommand struct {
	common.Command
}

func NewRemoveSalesCommand() *cobra.Command {
	var command *removeSalesCommand

	command = &removeSalesCommand{
		Command: common.Command{
			CobraCommand: &cobra.Command{
				Use:   "remove",
				Short: "Removes all sales",
				Long: heredoc.Doc(`
							This command remove the listed sales from the database.
							Use with caution, as this action cannot be undone.
						`),
				RunE: func(cmd *cobra.Command, args []string) error {
					return command.execute(args)
				},
				Args: cobra.MinimumNArgs(1),
			},
		},
	}

	return command.AsCobraCommand()
}

func (c *removeSalesCommand) execute(args []string) error {
	saleIDs, err := c.ParseSaleIDs(args)
	if err != nil {
		c.PrintErrorf("Failed to parse sale ids: %v", err)
	}

	transactionErr := c.WithTransaction(func(transaction *queries.TransactionalDatabaseQuerier) error {
		for _, saleID := range saleIDs {
			c.Printf("Removing sale with id %s\n", saleID.String())
			err := queries.RemoveSale(transaction, saleID)
			if err != nil {
				c.PrintErrorf("Failed to remove sale with id %s\n", saleID.String())
				return err
			}
		}

		return nil
	})

	if transactionErr != nil {
		c.PrintError(heredoc.Doc(`
			Row deletion took place inside transaction.
			An error occurred, causing a rollback.
			NONE of the sales was actually removed!
		`))
		return transactionErr
	}

	return nil
}
