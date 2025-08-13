package server

import (
	"bctbackend/algorithms"
	"bctbackend/commands/common"
	"bctbackend/server"
	"bctbackend/server/configuration"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
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
				Args: cobra.NoArgs,
			},
		},
	}

	command.CobraCommand.Flags().Int(common.CLIFlagPort, 8000, "Port to run the server on")
	command.CobraCommand.Flags().Bool(common.CLIFlagDebug, false, "Run server in debug mode")
	command.CobraCommand.Flags().String(common.CLIFlagHTML, "index.html", "Path to the HTML file to serve")
	viper.BindPFlag(common.ConfigKeyPort, command.CobraCommand.Flags().Lookup(common.CLIFlagPort))
	viper.BindPFlag("debug", command.CobraCommand.Flags().Lookup(common.CLIFlagDebug))
	viper.BindPFlag("html", command.CobraCommand.Flags().Lookup(common.CLIFlagHTML))

	return command.AsCobraCommand()
}

func (c *ServerCommand) execute() error {
	configuration, err := c.loadConfiguration()
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

func (c *ServerCommand) loadConfiguration() (*configuration.Configuration, error) {
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

	labelGeneration, err := c.getLabelGenerationConfiguration()
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

	configuration := configuration.Configuration{
		LogFilename:                 logFilename,
		LabelGeneration:             labelGeneration,
		Port:                        port,
		GinMode:                     ginMode,
		HTMLPath:                    htmlPath,
		ExpiredSessionPruneInterval: expiredSessionPruningInterval,
	}

	slog.Info("Loaded configuration successfully", "configuration", configuration.String())

	return &configuration, nil
}

func (c *ServerCommand) getLabelGenerationConfiguration() (*configuration.LabelGenerationConfiguration, error) {
	errs := []error{}

	barcodeWidth, err := c.GetConfigurationInt(common.FlagBarcodeWidth)
	if err != nil {
		errs = append(errs, err)
	}

	barcodeHeight, err := c.GetConfigurationInt(common.FlagBarcodeHeight)
	if err != nil {
		errs = append(errs, err)
	}

	font, err := c.getLabelFontConfiguration()
	if err != nil {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return nil, fmt.Errorf("failed to get label generation configuration: %w", errors.Join(errs...))
	}

	labelGeneration := configuration.LabelGenerationConfiguration{
		BarcodeWidth:  barcodeWidth,
		BarcodeHeight: barcodeHeight,
		Font:          font,
	}

	return &labelGeneration, nil
}

func (c *ServerCommand) getLabelFontConfiguration() (*configuration.FontConfiguration, error) {
	errs := []error{}

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

	if len(errs) > 0 {
		return nil, fmt.Errorf("failed to get font configuration: %w", errors.Join(errs...))
	}

	fontConfiguration := configuration.FontConfiguration{
		Directory: fontDirectory,
		Filename:  fontFilename,
		Family:    fontFamily,
	}

	return &fontConfiguration, nil
}

func (c *ServerCommand) ensureRequiredFilesExist(configuration *configuration.Configuration) error {
	fontPath := path.Join(configuration.LabelGeneration.Font.Directory, configuration.LabelGeneration.Font.Filename)
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
