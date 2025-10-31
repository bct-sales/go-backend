package commands

import (
	"bctbackend/commands/category"
	"bctbackend/commands/common"
	"bctbackend/commands/database"
	"bctbackend/commands/email"
	"bctbackend/commands/initialize"
	"bctbackend/commands/item"
	"bctbackend/commands/sale"
	"bctbackend/commands/server"
	"bctbackend/commands/session"
	"bctbackend/commands/user"
	"log/slog"
	"os"
	"strings"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewRootCommand() *cobra.Command {
	var verbose bool
	var noColor bool

	rootCommand := cobra.Command{
		Use:   "bctbackend",
		Short: "BCT Backend Command Line Interface",
		Long:  `BCT Backend Command Line Interface for managing items, users, and other resources.`,
	}

	cobra.OnInitialize(func() {
		if verbose {
			slog.SetLogLoggerLevel(slog.LevelDebug)
			slog.Info("Verbose mode enabled")
		}

		if noColor {
			pterm.DisableColor()
		}

		configurationPath := rootCommand.PersistentFlags().Lookup("config").Value.String()

		if configurationPath == "" {
			// Load configuration from file
			viper.SetConfigName("bctconfig")
			viper.SetConfigType("yaml")
			viper.AddConfigPath(".")
		} else {
			// Load configuration from specified file
			viper.SetConfigFile(configurationPath)
		}

		slog.Debug("Reading configuration")
		if err := viper.ReadInConfig(); err != nil {
			slog.Warn("Could not read configuration file", "error", err.Error())
		}
	})

	rootCommand.PersistentFlags().String("config", "", "Path to the configuration file")
	rootCommand.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
	rootCommand.PersistentFlags().BoolVar(&noColor, "no-color", false, "Disable color output")
	rootCommand.PersistentFlags().String("db", "./bct.db", "Path to the database file")

	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.SetEnvPrefix("BCT")
	viper.AutomaticEnv()

	viper.BindPFlag(common.CLIFlagConfigurationPath, rootCommand.PersistentFlags().Lookup("config"))
	viper.BindPFlag(common.ConfigKeyDatabase, rootCommand.PersistentFlags().Lookup("db"))

	rootCommand.AddCommand(item.NewItemCommand())
	rootCommand.AddCommand(user.NewUserCommand())
	rootCommand.AddCommand(database.NewDatabaseCommand())
	rootCommand.AddCommand(sale.NewSaleCommand())
	rootCommand.AddCommand(server.NewServerCommand())
	rootCommand.AddCommand(category.NewCategoryCommand())
	rootCommand.AddCommand(initialize.NewInitializeCommand())
	rootCommand.AddCommand(session.NewSessionCommand())
	rootCommand.AddCommand(email.NewEmailCommand())
	rootCommand.AddCommand(NewVersionCommand())
	rootCommand.AddCommand(NewUserInterfaceCommand())

	return &rootCommand
}

func Execute() {
	rootCommand := NewRootCommand()

	rootCommand.SilenceUsage = true
	// rootCommand.SilenceErrors = true

	if err := rootCommand.Execute(); err != nil {
		slog.Debug("An error occurred", "error", err.Error())
		os.Exit(1)
	}
}
