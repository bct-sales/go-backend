package common

import (
	"bctbackend/database/models"
	"database/sql"
	"fmt"

	"github.com/spf13/cobra"
)

type Command struct {
	CobraCommand *cobra.Command
}

func (c *Command) PrintErrorf(formatString string, args ...any) {
	fmt.Fprintf(c.CobraCommand.ErrOrStderr(), formatString, args...)
}

func (c *Command) Printf(formatString string, args ...any) {
	fmt.Fprintf(c.CobraCommand.OutOrStdout(), formatString, args...)
}

func (c *Command) WithOpenedDatabase(fn func(db *sql.DB) error) error {
	return WithOpenedDatabase(c.CobraCommand.ErrOrStderr(), fn)
}

func (c *Command) AsCobraCommand() *cobra.Command {
	return c.CobraCommand
}

func (c *Command) ParseItemId(str string) (models.Id, error) {
	return c.parseId(str, "item")
}

func (c *Command) ParseUserId(str string) (models.Id, error) {
	return c.parseId(str, "user")
}

func (c *Command) ParseSaleId(str string) (models.Id, error) {
	return c.parseId(str, "sale")
}

func (c *Command) parseId(str string, idType string) (models.Id, error) {
	id, err := models.ParseId(str)

	if err != nil {
		c.PrintErrorf("Invalid %s ID: %s\n", idType, str)
		return 0, err
	}

	return id, nil
}
