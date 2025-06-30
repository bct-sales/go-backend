package category

import (
	"bctbackend/commands/common"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"database/sql"
	"fmt"

	"github.com/spf13/cobra"
)

type addCategoryCommand struct {
	common.Command
	name string
	id   uint64
}

func NewCategoryAddCommand() *cobra.Command {
	var command *addCategoryCommand

	command = &addCategoryCommand{
		Command: common.Command{
			CobraCommand: &cobra.Command{
				Use:   "add",
				Short: "Add a new item category",
				Long:  `This command adds a new item category to the database.`,
				RunE: func(cmd *cobra.Command, args []string) error {
					return command.execute()
				},
			},
		},
	}

	command.CobraCommand.Flags().Uint64Var(&command.id, "id", 0, "ID of the new category")
	command.CobraCommand.Flags().StringVar(&command.name, "name", "", "Name of the new category")
	command.CobraCommand.MarkFlagRequired("id")
	command.CobraCommand.MarkFlagRequired("name")

	return command.AsCobraCommand()
}

func (c *addCategoryCommand) execute() error {
	return c.WithOpenedDatabase(func(database *sql.DB) error {
		if err := queries.AddCategoryWithId(database, models.Id(c.id), c.name); err != nil {
			return fmt.Errorf("failed to add category to database: %w", err)
		}

		c.Printf("Category added successfully")

		return nil
	})
}
