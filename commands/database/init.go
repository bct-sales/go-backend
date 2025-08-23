package database

import (
	"bctbackend/commands/common"
	db "bctbackend/database"
	dberr "bctbackend/database/errors"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"errors"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"
)

type InitializeDatabaseCommand struct {
	common.Command
	noCategories bool `exhaustruct:"optional"`
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
				Args: cobra.NoArgs,
			},
		},
	}

	command.CobraCommand.Flags().BoolVar(&command.noCategories, "no-categories", false, "Do not add default categories")

	return command.AsCobraCommand()
}

func (c *InitializeDatabaseCommand) execute() (r_err error) {
	databasePath, err := common.GetDatabasePath()
	if err != nil {
		c.PrintErrorf("Failed to get database path: %s\n", err.Error())
		return err
	}

	database, err := db.CreateDatabase(databasePath)

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

	defer func() {
		if err := database.Close(); err != nil {
			c.PrintErrorf("Failed to close database %s\n", databasePath)
			r_err = errors.Join(r_err, err)
		}
	}()

	if err := db.InitializeDatabase(database); err != nil {
		c.PrintErrorf("Failed to initialize database: %v\n", err)
		return err
	}

	if !c.noCategories {
		err := common.GenerateDefaultCategories(func(id models.Id, name string) error {
			return queries.AddCategoryWithId(database, id, name)
		})

		if err != nil {
			c.PrintErrorf("Failed to add default categories: %v\n", err)
			return err
		}
	}

	c.Printf("Database file successfully created at %s\n", databasePath)
	return nil
}
