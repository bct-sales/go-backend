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

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type ErrCommand struct {
	wrapped error
}

var ErrMissingConfiguration = errors.New("missing configuration")

func (e *ErrCommand) Error() string {
	if e.wrapped != nil {
		return "command error: " + e.wrapped.Error()
	}
	return "command error"
}

func (e *ErrCommand) Unwrap() error {
	return e.wrapped
}

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
		return &ErrCommand{wrapped: err}
	}

	db, err := database.OpenDatabase(databasePath)
	if err != nil {
		c.PrintErrorf("Failed to open database %s\n", databasePath)
		return &ErrCommand{wrapped: err}
	}

	defer func() {
		if err := db.Close(); err != nil {
			c.PrintErrorf("Failed to close database %s\n", databasePath)
			r_err = errors.Join(r_err, err)
		}
	}()

	return fn(db)
}

func (c *Command) WithTransaction(fn func(db *queries.TransactionalDatabaseQuerier) error) error {
	return c.WithOpenedDatabase(func(db *sql.DB) (r_err error) {
		transaction, err := queries.NewTransactionDatabaseQuerier(db)
		if err != nil {
			c.PrintErrorf("Failed to start transaction: %s\n", err.Error())
			return &ErrCommand{wrapped: err}
		}
		defer func() {
			if rollbackErr := transaction.Rollback(); rollbackErr != nil {
				c.PrintErrorf("Failed to roll back transaction: %s\n", rollbackErr.Error())
				r_err = errors.Join(r_err, rollbackErr)
			}
		}()

		if err := fn(transaction); err != nil {
			return fmt.Errorf("transaction failed: %w", err)
		}

		if err := transaction.Commit(); err != nil {
			return fmt.Errorf("failed to commit transaction: %w", err)
		}

		return nil
	})
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

func (c *Command) GetConfigurationInt(key string) (int, error) {
	if !viper.IsSet(key) {
		c.PrintErrorf("Configuration key '%s' is not set\n", key)
		return 0, fmt.Errorf("configuration key '%s' is not set", key)
	}

	value := viper.GetInt(key)
	slog.Debug("GetConfigurationInt", "key", key, "value", value)
	return value, nil
}

func (c *Command) GetConfigurationString(key string) (string, error) {
	slog.Debug("GetConfigurationString", "key", key, "isSet", viper.IsSet(key))

	if !viper.IsSet(key) {
		c.PrintErrorf("Configuration key '%s' is not set\n", key)
		return "", fmt.Errorf("configuration key '%s' is not set: %w", key, ErrMissingConfiguration)
	}

	value := viper.GetString(key)
	slog.Debug("GetConfigurationString", "key", key, "value", value)
	return viper.GetString(key), nil
}

func (c *Command) GetConfigurationBool(key string) (bool, error) {
	if !viper.IsSet(key) {
		c.PrintErrorf("Configuration key '%s' is not set\n", key)
		return false, fmt.Errorf("configuration key '%s' is not set: %w", key, ErrMissingConfiguration)
	}

	value := viper.GetBool(key)
	slog.Debug("GetConfigurationBool", "key", key, "value", value)
	return value, nil
}
