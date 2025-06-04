package commands

import (
	"bctbackend/commands/common"
	"bctbackend/commands/item"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/MakeNowJust/heredoc"
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
		if err := viper.ReadInConfig(); err != nil {
			handleMissingConfigurationFile()
		}
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

func handleMissingConfigurationFile() {
	absolutePathOfConfigurationFile, err := filepath.Abs(viper.ConfigFileUsed())

	if err != nil {
		fmt.Fprintln(os.Stderr, "Error determining absolute path of configuration file:", err)
		os.Exit(1)
	}

	fmt.Print(heredoc.Docf(`
		I could not find the configuration file.
		This is where I looked for it: %s.

		Possible solutions:
		* Specify a different path using the --config flag:
		    $ bctbackend --config path/to/your/config.yaml ...
		* You can also set the BCT_CONFIG environment variable to point to your configuration file.
		    $ BCT_CONFIG=path/to/your/config.yaml bctbackend ...
		* You can create a configuration file in the current directory with the name "bctconfig.yaml".
		    $ touch bctconfig.yaml
			$ bctbackend ...
		`,
		absolutePathOfConfigurationFile))
	os.Exit(1)
}
