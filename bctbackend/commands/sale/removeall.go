package sale

import (
	"bctbackend/commands/common"
	"bctbackend/database/queries"
	"database/sql"
	"fmt"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"
)

type removeAllSalesCommand struct {
	common.Command
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
			},
		},
	}

	return command.AsCobraCommand()
}

func (c *removeAllSalesCommand) execute() error {
	if err := c.EnsureConfigurationFileLoaded(); err != nil {
		return err
	}

	return c.WithOpenedDatabase(func(db *sql.DB) error {
		err := queries.RemoveAllSales(db)
		if err != nil {
			c.PrintErrorf("Failed to remove all sales\n")
			return fmt.Errorf("failed to remove all sales: %w", err)
		}

		c.Printf("All sales removed successfully.\n")
		return nil
	})
}
