package database

import (
	"github.com/spf13/cobra"
)

func NewDatabaseCommand() *cobra.Command {
	command := cobra.Command{
		Use:   "db",
		Short: "Performs database-level operations",
		Long:  `Commands to perform operations on database level.`,
	}

	command.AddCommand(NewDatabaseBackupCommand())
	command.AddCommand(NewDatabaseInitCommand())
	command.AddCommand(NewDatabaseDummyCommand())

	return &command
}
