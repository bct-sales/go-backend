package server

import (
	"bctbackend/algorithms"
	"bctbackend/clock"
	"bctbackend/commands/common"
	"bctbackend/logging"
	"bctbackend/server"
	"bctbackend/server/configuration"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/natefinch/lumberjack.v2"
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
	viper.BindPFlag(common.ConfigKeyDebug, command.CobraCommand.Flags().Lookup(common.CLIFlagDebug))
	viper.BindPFlag(common.ConfigKeyHTML, command.CobraCommand.Flags().Lookup(common.CLIFlagHTML))

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

	clock := clock.NewSystemClock()

	logger, loggerErr := c.createLogger(configuration.Log)
	if loggerErr != nil {
		c.PrintErrorf("Failed to create logger\n")
		return fmt.Errorf("failed to create logger: %w", err)
	}

	return c.WithOpenedDatabase(func(db *sql.DB) error {
		if err := server.StartServer(clock, db, logger, configuration); err != nil {
			c.PrintErrorf("Failed to start REST service\n")
			return fmt.Errorf("failed to start REST service: %w", err)
		}

		return nil
	})
}

func (c *ServerCommand) loadConfiguration() (*configuration.Configuration, error) {
	errs := []error{}

	labelGenerationConfiguration, err := c.getLabelGenerationConfiguration()
	if err != nil {
		errs = append(errs, err)
	}

	logConfiguration, err := c.getLogConfiguration()
	if err != nil {
		errs = append(errs, err)
	}

	serverConfiguration, err := c.getServerConfiguration()
	if err != nil {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}

	configuration := configuration.Configuration{
		Log:             logConfiguration,
		LabelGeneration: labelGenerationConfiguration,
		Server:          serverConfiguration,
	}

	slog.Info("Loaded configuration successfully", "configuration", configuration.String())

	return &configuration, nil
}

func (c *ServerCommand) getLogConfiguration() (*configuration.LogConfiguration, error) {
	errs := []error{}

	logFile, err := c.GetConfigurationString(common.ConfigKeyLogFile)
	if err != nil {
		if errors.Is(err, common.ErrMissingConfiguration) {
			return nil, nil
		} else {
			errs = append(errs, err)
		}
	}

	logMaxSize, err := c.GetConfigurationInt(common.ConfigKeyLogMaxSize)
	if err != nil {
		errs = append(errs, err)
	}

	logMaxBackups, err := c.GetConfigurationInt(common.ConfigKeyLogMaxBackups)
	if err != nil {
		errs = append(errs, err)
	}

	logMaxAge, err := c.GetConfigurationInt(common.ConfigKeyLogMaxAge)
	if err != nil {
		errs = append(errs, err)
	}

	logCompression, err := c.GetConfigurationBool(common.ConfigKeyLogCompression)
	if err != nil {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return nil, fmt.Errorf("failed to get log configuration: %w", errors.Join(errs...))
	}

	logConfiguration := configuration.LogConfiguration{
		File:             logFile,
		MaxSizeMegabytes: logMaxSize,
		MaxBackups:       logMaxBackups,
		MaxAgeDays:       logMaxAge,
		Compression:      logCompression,
	}
	return &logConfiguration, nil
}

func (c *ServerCommand) createLogger(configuration *configuration.LogConfiguration) (logging.Logger, error) {
	var writer io.Writer

	if configuration == nil {
		writer = os.Stderr
	} else {
		//exhaustruct:ignore
		loggerFile := lumberjack.Logger{
			Filename:   configuration.File,
			MaxSize:    configuration.MaxSizeMegabytes,
			MaxBackups: configuration.MaxBackups,
			MaxAge:     configuration.MaxAgeDays,
			Compress:   configuration.Compression,
		}

		writer = io.MultiWriter(os.Stderr, &loggerFile)
	}

	slogger := slog.New(slog.NewJSONHandler(writer, nil))
	logger := logging.NewSloggerWrapper(slogger)

	return logger, nil
}

func (c *ServerCommand) getServerConfiguration() (*configuration.ServerConfiguration, error) {
	errs := []error{}

	expiredSessionPruningInterval, err := c.GetConfigurationInt(common.ConfigKeyPruneExpiredSessionsInterval)
	if err != nil {
		errs = append(errs, err)
	}

	htmlPath, err := c.GetConfigurationString(common.ConfigKeyHTML)
	if err != nil {
		errs = append(errs, err)
	}

	port, err := c.GetConfigurationInt(common.ConfigKeyPort)
	if err != nil {
		errs = append(errs, err)
	}

	debugMode, err := c.GetConfigurationBool(common.ConfigKeyDebug)
	if err != nil {
		errs = append(errs, err)
	}

	cookieDomain, err := c.GetConfigurationString(common.ConfigKeyCookieDomain)
	if err != nil {
		errs = append(errs, err)
	}

	swagger, err := c.GetConfigurationBool(common.ConfigKeySwagger)
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

	serverConfiguration := configuration.ServerConfiguration{
		Port:                        port,
		GinMode:                     ginMode,
		HTMLPath:                    htmlPath,
		ExpiredSessionPruneInterval: expiredSessionPruningInterval,
		CookieDomain:                cookieDomain,
		Swagger:                     swagger,
	}

	if len(errs) > 0 {
		return nil, fmt.Errorf("failed to get server configuration: %w", errors.Join(errs...))
	}

	return &serverConfiguration, nil
}

func (c *ServerCommand) getLabelGenerationConfiguration() (*configuration.LabelGenerationConfiguration, error) {
	errs := []error{}

	barcodeWidth, err := c.GetConfigurationInt(common.ConfigKeyLabelBarcodeWidth)
	if err != nil {
		errs = append(errs, err)
	}

	barcodeHeight, err := c.GetConfigurationInt(common.ConfigKeyLabelBarcodeHeight)
	if err != nil {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return nil, fmt.Errorf("failed to get label generation configuration: %w", errors.Join(errs...))
	}

	labelGeneration := configuration.LabelGenerationConfiguration{
		BarcodeWidth:  barcodeWidth,
		BarcodeHeight: barcodeHeight,
	}

	return &labelGeneration, nil
}

func (c *ServerCommand) ensureRequiredFilesExist(configuration *configuration.Configuration) error {
	if err := c.ensureFileExists(configuration.Server.HTMLPath); err != nil {
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
