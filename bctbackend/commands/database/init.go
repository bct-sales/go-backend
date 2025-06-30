package database

import (
	"bctbackend/commands/common"
	"bctbackend/database"
	dberr "bctbackend/database/errors"
	"errors"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"
)

type InitializeDatabaseCommand struct {
	common.Command
	noCategories bool
}

func NewDatabaseInitCommand() *cobra.Command {
	var command *InitializeDatabaseCommand

	command = &InitializeDatabaseCommand{
		Command: common.Command{
			CobraCommand: &cobra.Command{
				Use:   "init",
				Short: "Initialize the database",
				Long: heredoc.Doc(`
					This command makes creates a new database file and initializes the database.
					It will also create the default categories unless the --no-categories flag is set.
					If a database file already exists, it will NOT be overwritten.
					If you need to create a fresh database, either use a different path
					or delete the existing database.
				`),
				RunE: func(cmd *cobra.Command, args []string) error {
					return command.execute()
				},
			},
		},
	}

	command.CobraCommand.Flags().BoolVar(&command.noCategories, "no-categories", false, "Do not add default categories")

	return command.AsCobraCommand()
}

func (c *InitializeDatabaseCommand) execute() error {
	databasePath, err := common.GetDatabasePath()
	if err != nil {
		c.PrintErrorf("Failed to get database path: %s\n", err.Error())
		return err
	}

	db, err := database.CreateDatabase(databasePath)

	if err != nil {
		if errors.Is(err, dberr.ErrDatabaseAlreadyExists) {
			c.PrintErrorf(heredoc.Docf(
				`
					Database file %s already exists.
					To create a new database, either use a different path or delete the existing file.
				`, databasePath))

			return err
		}

		c.PrintErrorf("Failed to create database file: %v\n", err)
		return err
	}

	defer db.Close()

	c.Printf("Database file successfully created at %s\n", databasePath)
	return nil
}
