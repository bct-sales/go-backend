package common

import (
	"bctbackend/database"
	"database/sql"
	"errors"
	"fmt"
	"io"

	"github.com/spf13/viper"
)

func GetDatabasePath() (string, error) {
	if !viper.IsSet(ConfigKeyDatabase) {
		return "", fmt.Errorf("database path is not set in configuration")
	}

	return viper.GetString(ConfigKeyDatabase), nil
}

func WithOpenedDatabase(writer io.Writer, callback func(db *sql.DB) error) (err error) {
	databasePath, err := GetDatabasePath()
	if err != nil {
		fmt.Fprintf(writer, "Failed to get database path: %s\n", err.Error())
		return err
	}

	database, err := database.OpenDatabase(databasePath)
	if err != nil {
		fmt.Fprintf(writer, "Failed to open database %s\n", databasePath)
		return
	}

	defer func() {
		if errClose := database.Close(); errClose != nil {
			fmt.Fprintf(writer, "Failed to close database %s\n", databasePath)
			err = errors.Join(err, errClose)
		}
	}()

	return callback(database)
}
