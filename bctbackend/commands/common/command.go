package common

import (
	"bctbackend/algorithms"
	"bctbackend/database"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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

func (c *Command) LoadConfigurationFile() error {
	absolutePathOfConfigurationFile, err := filepath.Abs(viper.ConfigFileUsed())

	if err != nil {
		fmt.Fprintln(os.Stderr, "Error determining absolute path of configuration file:", err)
		os.Exit(1)
	}

	slog.Debug("Reading configuration")
	if err := viper.ReadInConfig(); err != nil {
		fmt.Print(heredoc.Docf(
			`
				I could not find the configuration file.
				I looked for it here: %s.

				Possible solutions:
				* Use this tool to generate a configuration file:
					$ bctbackend init
				after which you can edit the configuration file.
				* Specify a different path using the --config flag:
					$ bctbackend --config path/to/your/config.yaml ...
				* You can also set the BCT_CONFIG environment variable to point to your configuration file.
					$ BCT_CONFIG=path/to/your/config.yaml bctbackend ...
			`,
			absolutePathOfConfigurationFile))

		return err
	}

	return nil
}

func (c *Command) EnsureConfigurationFileLoaded() error {
	if c.LoadConfigurationFile() != nil {
		c.PrintErrorf("Failed to load configuration file.\n")
		return fmt.Errorf("failed to load configuration file")
	}

	return nil
}
