package commands

import (
	"bctbackend/commands/common"
	"bctbackend/commands/item"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewRootCommand() *cobra.Command {
	var verbose bool

	rootCommand := cobra.Command{
		Use:   "bctbackend",
		Short: "BCT Backend Command Line Interface",
		Long:  `BCT Backend Command Line Interface for managing items, users, and other resources.`,
	}

	cobra.OnInitialize(func() {
		if verbose {
			slog.SetLogLoggerLevel(slog.LevelDebug)
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
		cobra.CheckErr(viper.ReadInConfig())
		slog.Debug("Done reading configuration", "file", viper.ConfigFileUsed())
	})

	rootCommand.PersistentFlags().String("config", "", "Path to the configuration file")
	rootCommand.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
	rootCommand.PersistentFlags().String("db", "./bct.db", "Path to the database file")

	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.SetEnvPrefix("BCT")
	viper.AutomaticEnv()

	viper.BindPFlag(common.FlagConfigurationPath, rootCommand.PersistentFlags().Lookup("config"))
	viper.BindPFlag(common.FlagDatabase, rootCommand.PersistentFlags().Lookup("db"))
	viper.SetDefault("barcode.width", 150)
	viper.SetDefault("barcode.height", 30)

	rootCommand.AddCommand(item.NewItemCommand())

	return &rootCommand
}

func Execute() {
	rootCommand := NewRootCommand()

	if err := rootCommand.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
