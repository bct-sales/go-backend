package sale

import (
	"bctbackend/commands/common"
	"bctbackend/database/queries"
	"fmt"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"
)

type removeAllSalesCommand struct {
	common.Command
	force bool `exhaustruct:"optional"`
}

func NewRemoveAllSalesCommand() *cobra.Command {
	var command *removeAllSalesCommand

	command = &removeAllSalesCommand{
		Command: common.Command{
			CobraCommand: &cobra.Command{
				Use:   "remove-all",
				Short: "Removes all sales",
				Long: heredoc.Doc(`
							This command removes all sales from the database.
							Use with caution, as this action cannot be undone.
						`),
				RunE: func(cmd *cobra.Command, args []string) error {
					return command.execute()
				},
				Args: cobra.NoArgs,
			},
		},
	}

	command.CobraCommand.Flags().BoolVar(&command.force, "force", false, "Required")

	return command.AsCobraCommand()
}

func (c *removeAllSalesCommand) execute() error {
	if !c.force {
		c.Printf(heredoc.Doc(`
								This action is irreversible and will delete all sales from the database!
								Please use --force to indicate you are aware of the consequences")
							`))
		return nil
	}
	return c.WithTransaction(func(transaction *queries.TransactionalDatabaseQuerier) error {
		err := queries.RemoveAllSales(transaction)
		if err != nil {
			c.PrintErrorf("Failed to remove all sales\n")
			return fmt.Errorf("failed to remove all sales: %w", err)
		}

		c.Printf("All sales removed successfully.\n")
		return nil
	})
}
