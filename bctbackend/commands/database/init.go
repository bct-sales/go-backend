package database

import (
	"bctbackend/commands/common"
	"bctbackend/database"
	dberr "bctbackend/database/errors"
	"errors"
	"fmt"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"
)

func NewDatabaseInitCommand() *cobra.Command {
	var noCategories bool

	command := cobra.Command{
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
			databasePath, err := common.GetDatabasePath()
			if err != nil {
				// TODO Write to correct output stream
				fmt.Fprintf(cmd.ErrOrStderr(), "Failed to get database path: %s\n", err.Error())
				return err
			}

			db, err := database.CreateDatabase(databasePath)

			if err != nil {
				if errors.Is(err, dberr.ErrDatabaseAlreadyExists) {
					fmt.Fprint(cmd.
						ErrOrStderr(),
						heredoc.Docf(`
							Database file %s already exists.
							To create a new database, either use a different path or delete the existing file.
						`, databasePath))
					return err
				}
				fmt.Fprintf(cmd.ErrOrStderr(), "Failed to create database file: %v\n", err)
				return err
			}

			db.Close()

			fmt.Fprintf(cmd.OutOrStdout(), "Database file successfully created at %s\n", databasePath)
			return nil
		},
	}

	command.Flags().BoolVar(&noCategories, "no-categories", false, "Do not add default categories")

	return &command
}
