package common

import (
	"database/sql"
	"fmt"

	"github.com/spf13/cobra"
)

type Command struct {
	CobraCommand *cobra.Command
}

func (c *Command) PrintError(formatString string, args ...any) {
	fmt.Fprintf(c.CobraCommand.ErrOrStderr(), formatString, args...)
}

func (c *Command) Printf(formatString string, args ...any) {
	fmt.Fprintf(c.CobraCommand.OutOrStdout(), formatString, args...)
}

func (c *Command) WithOpenedDatabase(fn func(db *sql.DB) error) error {
	return WithOpenedDatabase(c.CobraCommand.ErrOrStderr(), fn)
}
