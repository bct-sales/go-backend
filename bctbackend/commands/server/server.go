package server

import (
	"bctbackend/algorithms"
	"bctbackend/commands/common"
	"bctbackend/server"
	"bctbackend/server/configuration"
	"database/sql"
	"errors"
	"fmt"
	"path"

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
	command.CobraCommand.Flags().String("html", "index.html", "Path to the HTML file to serve")
	viper.BindPFlag("port", command.CobraCommand.Flags().Lookup("port"))
	viper.BindPFlag("debug", command.CobraCommand.Flags().Lookup("debug"))
	viper.BindPFlag("html", command.CobraCommand.Flags().Lookup("html"))

	return command.AsCobraCommand()
}

func (c *ServerCommand) execute() error {
	configuration, err := c.getConfiguration()
	if err != nil {
		return err
	}

	if err := c.ensureRequiredFilesExist(configuration); err != nil {
		return err
	}

	return c.WithOpenedDatabase(func(db *sql.DB) error {
		if err := server.StartServer(db, configuration); err != nil {
			c.PrintErrorf("Failed to start REST service\n")
			return fmt.Errorf("failed to start REST service: %w", err)
		}

		return nil
	})
}

func (c *ServerCommand) getConfiguration() (*configuration.Configuration, error) {
	errs := []error{}

	var logFilename *string
	logFileSetting, err := c.GetConfigurationString("log_file")
	if err != nil {
		if errors.Is(err, common.ErrMissingConfiguration) {
			logFilename = nil
		} else {
			return nil, fmt.Errorf("failed to get log_file configuration: %w", err)
		}
	}
	logFilename = &logFileSetting

	fontDirectory, err := c.GetConfigurationString(common.FlagFontDirectory)
	if err != nil {
		errs = append(errs, err)
	}

	fontFilename, err := c.GetConfigurationString(common.FlagFontFilename)
	if err != nil {
		errs = append(errs, err)
	}

	fontFamily, err := c.GetConfigurationString(common.FlagFontFamily)
	if err != nil {
		errs = append(errs, err)
	}

	barcodeWidth, err := c.GetConfigurationInt(common.FlagBarcodeWidth)
	if err != nil {
		errs = append(errs, err)
	}

	barcodeHeight, err := c.GetConfigurationInt(common.FlagBarcodeHeight)
	if err != nil {
		errs = append(errs, err)
	}

	port, err := c.GetConfigurationInt("port")
	if err != nil {
		errs = append(errs, err)
	}

	debugMode, err := c.GetConfigurationBool("debug")
	if err != nil {
		errs = append(errs, err)
	}

	expiredSessionPruningInterval, err := c.GetConfigurationInt("expired_session_prune_interval")
	if err != nil {
		errs = append(errs, err)
	}

	htmlPath, err := c.GetConfigurationString("html")
	if err != nil {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return nil, fmt.Errorf("failed to get configuration: %w", errors.Join(errs...))
	}

	var ginMode string
	if debugMode {
		ginMode = "debug"
	} else {
		ginMode = "release"
	}

	return &configuration.Configuration{
		LogFilename:                 logFilename,
		FontDirectory:               fontDirectory,
		FontFilename:                fontFilename,
		FontFamily:                  fontFamily,
		BarcodeWidth:                barcodeWidth,
		BarcodeHeight:               barcodeHeight,
		Port:                        port,
		GinMode:                     ginMode,
		HTMLPath:                    htmlPath,
		ExpiredSessionPruneInterval: expiredSessionPruningInterval,
	}, nil
}

func (c *ServerCommand) ensureRequiredFilesExist(configuration *configuration.Configuration) error {
	fontPath := path.Join(configuration.FontDirectory, configuration.FontFilename)
	if err := c.ensureFileExists(fontPath); err != nil {
		return fmt.Errorf("failed while checking font file existence: %w", err)
	}

	if err := c.ensureFileExists(configuration.HTMLPath); err != nil {
		return fmt.Errorf("failed while checking for html file existence: %w", err)
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
