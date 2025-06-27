package rest

import (
	"bctbackend/commands/common"
	"bctbackend/server"
	"bctbackend/server/configuration"
	"database/sql"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type ServerCommand struct {
	common.Command
}

func NewServerCommand() *cobra.Command {
	var command *ServerCommand

	command = &ServerCommand{
		Command: common.Command{
			CobraCommand: &cobra.Command{
				Use:   "server",
				Short: "Start server",
				Long:  `This command starts the backend server for the BCT application.`,
				RunE: func(cmd *cobra.Command, args []string) error {
					return command.execute()
				},
			},
		},
	}

	return command.AsCobraCommand()
}

func (c *ServerCommand) execute() error {
	configuration := configuration.Configuration{
		FontDirectory: viper.GetString(common.FlagFontDirectory),
		FontFilename:  viper.GetString(common.FlagFontFilename),
		FontFamily:    viper.GetString(common.FlagFontFamily),
		BarcodeWidth:  viper.GetInt(common.FlagBarcodeWidth),
		BarcodeHeight: viper.GetInt(common.FlagBarcodeHeight),
	}

	return c.WithOpenedDatabase(func(db *sql.DB) error {
		if err := server.StartServer(db, &configuration); err != nil {
			c.PrintErrorf("Failed to start REST service\n")
			return fmt.Errorf("failed to start REST service: %w", err)
		}

		return nil
	})
}
