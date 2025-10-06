package category

import (
	"bctbackend/commands/common"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"database/sql"
	"fmt"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"
)

type renameCategoryCommand struct {
	common.Command
	id      uint64 `exhaustruct:"optional"`
	newName string `exhaustruct:"optional"`
}

func NewRenameCategoryCommand() *cobra.Command {
	var command *renameCategoryCommand

	command = &renameCategoryCommand{
		Command: common.Command{
			CobraCommand: &cobra.Command{
				Use:   "rename",
				Short: "Rename an item category",
				Long: heredoc.Doc(`
							This command renames a category.
						`),
				Args: cobra.NoArgs,
				RunE: func(cmd *cobra.Command, args []string) error {
					return command.execute()
				},
			},
		},
	}

	command.CobraCommand.Flags().Uint64Var(&command.id, "id", 0, "ID of the category to be renamed")
	command.CobraCommand.Flags().StringVar(&command.newName, "name", "", "New name of the category")
	command.CobraCommand.MarkFlagRequired("id")
	command.CobraCommand.MarkFlagRequired("name")

	return command.AsCobraCommand()
}

func (c *renameCategoryCommand) execute() error {
	return c.WithOpenedDatabase(func(database *sql.DB) error {
		if err := queries.RenameCategory(database, models.ID(c.id), c.newName); err != nil {
			return fmt.Errorf("failed to rename category: %w", err)
		}

		c.Printf("Category successfully renamed\n")

		return nil
	})
}
