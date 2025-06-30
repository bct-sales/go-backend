package common

import (
	"bctbackend/algorithms"
	"bctbackend/database"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"database/sql"
	"errors"
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

func (c *Command) WithOpenedDatabase(fn func(db *sql.DB) error) (r_err error) {
	databasePath, err := GetDatabasePath()
	if err != nil {
		c.PrintErrorf("Failed to get database path: %s\n", err.Error())
		return err
	}

	db, err := database.OpenDatabase(databasePath)
	if err != nil {
		c.PrintErrorf("Failed to open database %s\n", databasePath)
		return err
	}

	defer func() {
		if err := db.Close(); err != nil {
			c.PrintErrorf("Failed to close database %s\n", databasePath)
			r_err = errors.Join(r_err, err)
		}
	}()

	return fn(db)
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

func (c *Command) ParseItemIds(str []string) ([]models.Id, error) {
	return algorithms.MapError(str, c.ParseItemId)
}

func (c *Command) ParseUserIds(str []string) ([]models.Id, error) {
	return algorithms.MapError(str, c.ParseUserId)
}

func (c *Command) ParseSaleIds(str []string) ([]models.Id, error) {
	return algorithms.MapError(str, c.ParseSaleId)
}

func (c *Command) parseId(str string, idType string) (models.Id, error) {
	id, err := models.ParseId(str)

	if err != nil {
		c.PrintErrorf("Invalid %s ID: %s\n", idType, str)
		return 0, err
	}

	return id, nil
}

func (c *Command) GetCategoryNameTable(db *sql.DB) (map[models.Id]string, error) {
	categoryNameTable, err := queries.GetCategoryNameTable(db)

	if err != nil {
		c.PrintErrorf("Failed to get category name table: %v\n", err)
		return nil, fmt.Errorf("failed to get category name table: %w", err)
	}

	return categoryNameTable, nil
}
