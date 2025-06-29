package server

import (
	"bctbackend/algorithms"
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

	command.CobraCommand.Flags().Int("port", 8000, "Port to run the server on")
	command.CobraCommand.Flags().Bool("debug", false, "Run server in debug mode")
	viper.BindPFlag("port", command.CobraCommand.Flags().Lookup("port"))
	viper.BindPFlag("debug", command.CobraCommand.Flags().Lookup("debug"))
	viper.SetDefault("port", 8000)
	viper.SetDefault("debug", false)

	return command.AsCobraCommand()
}

func (c *ServerCommand) execute() error {
	var ginMode string
	if viper.GetBool("debug") {
		ginMode = "debug"
	} else {
		ginMode = "release"
	}

	configuration := configuration.Configuration{
		FontDirectory: viper.GetString(common.FlagFontDirectory),
		FontFilename:  viper.GetString(common.FlagFontFilename),
		FontFamily:    viper.GetString(common.FlagFontFamily),
		BarcodeWidth:  viper.GetInt(common.FlagBarcodeWidth),
		BarcodeHeight: viper.GetInt(common.FlagBarcodeHeight),
		Port:          viper.GetInt("port"),
		GinMode:       ginMode,
	}

	if err := c.ensureRequiredFilesExist(&configuration); err != nil {
		return err
	}

	return c.WithOpenedDatabase(func(db *sql.DB) error {
		if err := server.StartServer(db, &configuration); err != nil {
			c.PrintErrorf("Failed to start REST service\n")
			return fmt.Errorf("failed to start REST service: %w", err)
		}

		return nil
	})
}

func (c *ServerCommand) ensureRequiredFilesExist(configuration *configuration.Configuration) error {
	if err := c.ensureFileExists(configuration.FontFilename); err != nil {
		return err
	}

	return nil
}

func (c *ServerCommand) ensureFileExists(path string) error {
	exists, err := algorithms.FileExists(path)

	if err != nil {
		c.PrintErrorf("Failed to check if file exists: %v\n", err)
		return fmt.Errorf("failed to check if file exists: %w", err)
	}

	if !exists {
		c.PrintErrorf("Required file does not exist: %s\n", path)
		return fmt.Errorf("required file does not exist: %s", path)
	}

	return nil
}
