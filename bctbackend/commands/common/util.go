package common

import (
	"bctbackend/database"
	"database/sql"
	"errors"
	"fmt"
	"io"

	"github.com/spf13/viper"
)

func GetDatabasePath() string {
	return viper.GetString(FlagDatabase)
}

func WithOpenedDatabase(writer io.Writer, fn func(db *sql.DB) error) (r_err error) {
	databasePath := GetDatabasePath()
	db, err := database.OpenDatabase(databasePath)
	if err != nil {
		fmt.Fprintf(writer, "Failed to open database %s\n", databasePath)
		return
	}

	defer func() {
		if err := db.Close(); err != nil {
			fmt.Fprintf(writer, "Failed to close database %s\n", databasePath)
			r_err = errors.Join(r_err, err)
		}
	}()

	return fn(db)
}
