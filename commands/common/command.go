package common

import (
	"bctbackend/algorithms"
	"bctbackend/database"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"context"
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

func (command *Command) PrintErrorf(formatString string, args ...any) {
	fmt.Fprintf(command.CobraCommand.ErrOrStderr(), formatString, args...)
}

func (command *Command) Printf(formatString string, args ...any) {
	fmt.Fprintf(command.CobraCommand.OutOrStdout(), formatString, args...)
}

func (command *Command) WithOpenedDatabase(callback func(db *sql.DB) error) (r_err error) {
	databasePath, err := GetDatabasePath()
	if err != nil {
		command.PrintErrorf("Failed to get database path: %s\n", err.Error())
		return &ErrCommand{wrapped: err}
	}

	database, err := database.OpenDatabase(databasePath)
	if err != nil {
		command.PrintErrorf("Failed to open database %s\n", databasePath)
		return &ErrCommand{wrapped: err}
	}

	defer func() {
		if err := database.Close(); err != nil {
			command.PrintErrorf("Failed to close database %s\n", databasePath)
			r_err = errors.Join(r_err, err)
		}
	}()

	return callback(database)
}

func (command *Command) WithTransaction(callback func(db *queries.TransactionalDatabaseQuerier) error) error {
	return command.WithOpenedDatabase(func(db *sql.DB) (r_err error) {
		transaction, err := queries.NewTransactionalDatabaseQuerier(context.Background(), db)
		if err != nil {
			command.PrintErrorf("Failed to start transaction: %s\n", err.Error())
			return &ErrCommand{wrapped: err}
		}
		defer transaction.RollbackIfNotCommitted()

		if err := callback(transaction); err != nil {
			return fmt.Errorf("transaction failed: %w", err)
		}

		if err := transaction.Commit(); err != nil {
			return fmt.Errorf("failed to commit transaction: %w", err)
		}

		return nil
	})
}

func (command *Command) AsCobraCommand() *cobra.Command {
	return command.CobraCommand
}

func (command *Command) ParseItemId(str string) (models.Id, error) {
	return command.parseId(str, "item")
}

func (command *Command) ParseUserId(str string) (models.Id, error) {
	return command.parseId(str, "user")
}

func (command *Command) ParseSaleId(str string) (models.Id, error) {
	return command.parseId(str, "sale")
}

func (command *Command) ParseItemIds(str []string) ([]models.Id, error) {
	return algorithms.MapError(str, command.ParseItemId)
}

func (command *Command) ParseUserIds(str []string) ([]models.Id, error) {
	return algorithms.MapError(str, command.ParseUserId)
}

func (command *Command) ParseSaleIds(str []string) ([]models.Id, error) {
	return algorithms.MapError(str, command.ParseSaleId)
}

func (command *Command) parseId(str string, idType string) (models.Id, error) {
	id, err := models.ParseId(str)

	if err != nil {
		command.PrintErrorf("Invalid %s ID: %s\n", idType, str)
		return 0, err
	}

	return id, nil
}

func (command *Command) GetCategoryNameTable(db queries.DatabaseQuerier) (map[models.Id]string, error) {
	categoryNameTable, err := queries.GetCategoryNameTable(db)

	if err != nil {
		command.PrintErrorf("Failed to get category name table: %v\n", err)
		return nil, fmt.Errorf("failed to get category name table: %w", err)
	}

	return categoryNameTable, nil
}

func (command *Command) GetConfigurationInt(key string) (int, error) {
	if !viper.IsSet(key) {
		command.PrintErrorf("Configuration key '%s' is not set\n", key)
		return 0, fmt.Errorf("configuration key '%s' is not set", key)
	}

	value := viper.GetInt(key)
	slog.Debug("GetConfigurationInt", "key", key, "value", value)
	return value, nil
}

func (command *Command) GetConfigurationString(key string) (string, error) {
	slog.Debug("GetConfigurationString", "key", key, "isSet", viper.IsSet(key))

	if !viper.IsSet(key) {
		command.PrintErrorf("Configuration key '%s' is not set\n", key)
		return "", fmt.Errorf("configuration key '%s' is not set: %w", key, ErrMissingConfiguration)
	}

	value := viper.GetString(key)
	slog.Debug("GetConfigurationString", "key", key, "value", value)
	return viper.GetString(key), nil
}

func (command *Command) GetConfigurationBool(key string) (bool, error) {
	if !viper.IsSet(key) {
		command.PrintErrorf("Configuration key '%s' is not set\n", key)
		return false, fmt.Errorf("configuration key '%s' is not set: %w", key, ErrMissingConfiguration)
	}

	value := viper.GetBool(key)
	slog.Debug("GetConfigurationBool", "key", key, "value", value)
	return value, nil
}
