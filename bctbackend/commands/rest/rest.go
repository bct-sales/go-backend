package rest

import (
	"bctbackend/commands/common"
	"bctbackend/server"
	"database/sql"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type RestCommand struct {
	common.Command
}

func NewRestCommand() *cobra.Command {
	var command *RestCommand

	command = &RestCommand{
		Command: common.Command{
			CobraCommand: &cobra.Command{
				Use:   "server",
				Short: "Start REST server",
				Long:  `This command starts the REST server for the BCT application.`,
				RunE: func(cmd *cobra.Command, args []string) error {
					return command.execute()
				},
			},
		},
	}

	return command.AsCobraCommand()
}

func (c *RestCommand) execute() error {
	configuration := server.Configuration{
		FontDirectory: viper.GetString(common.FlagFontDirectory),
		FontFilename:  viper.GetString(common.FlagFontFilename),
		FontFamily:    viper.GetString(common.FlagFontFamily),
		BarcodeWidth:  viper.GetInt(common.FlagBarcodeWidth),
		BarcodeHeight: viper.GetInt(common.FlagBarcodeHeight),
	}

	return c.WithOpenedDatabase(func(db *sql.DB) error {
		if err := server.StartRestService(db, &configuration); err != nil {
			c.PrintErrorf("Failed to start REST service\n")
			return fmt.Errorf("failed to start REST service: %w", err)
		}

		return nil
	})
}
