package commands

import (
	"bctbackend/commands/category"
	"bctbackend/commands/common"
	"bctbackend/commands/database"
	"bctbackend/commands/download"
	"bctbackend/commands/initialize"
	"bctbackend/commands/item"
	"bctbackend/commands/sale"
	"bctbackend/commands/server"
	"bctbackend/commands/user"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/MakeNowJust/heredoc"
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
	})

	rootCommand.PersistentFlags().String("config", "", "Path to the configuration file")
	rootCommand.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
	rootCommand.PersistentFlags().BoolVar(&noColor, "no-color", false, "Disable color output")
	rootCommand.PersistentFlags().String("db", "./bct.db", "Path to the database file")

	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.SetEnvPrefix("BCT")
	viper.AutomaticEnv()

	viper.BindPFlag(common.FlagConfigurationPath, rootCommand.PersistentFlags().Lookup("config"))
	viper.BindPFlag(common.FlagDatabase, rootCommand.PersistentFlags().Lookup("db"))
	viper.SetDefault(common.FlagFontDirectory, ".")
	viper.SetDefault(common.FlagFontFilename, "arial.ttf")
	viper.SetDefault(common.FlagFontFamily, "Arial")
	viper.SetDefault(common.FlagBarcodeWidth, 150)
	viper.SetDefault(common.FlagBarcodeHeight, 30)

	rootCommand.AddCommand(item.NewItemCommand())
	rootCommand.AddCommand(user.NewUserCommand())
	rootCommand.AddCommand(database.NewDatabaseCommand())
	rootCommand.AddCommand(sale.NewSaleCommand())
	rootCommand.AddCommand(server.NewServerCommand())
	rootCommand.AddCommand(category.NewCategoryCommand())
	rootCommand.AddCommand(initialize.NewInitializeCommand())
	rootCommand.AddCommand(download.NewDownloadCommand())

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

func handleMissingConfigurationFile() {
	absolutePathOfConfigurationFile, err := filepath.Abs(viper.ConfigFileUsed())

	if err != nil {
		fmt.Fprintln(os.Stderr, "Error determining absolute path of configuration file:", err)
		os.Exit(1)
	}

	fmt.Print(heredoc.Docf(`
		=======
		WARNING
		=======
		I could not find the configuration file.
		I looked for it here: %s.

		Possible solutions:
		* Use this tool to generate a configuration file:
			$ bctbackend init
		* Specify a different path using the --config flag:
		    $ bctbackend --config path/to/your/config.yaml ...
		* You can also set the BCT_CONFIG environment variable to point to your configuration file.
		    $ BCT_CONFIG=path/to/your/config.yaml bctbackend ...
		* You can create a configuration file in the current directory with the name "bctconfig.yaml".
		    $ touch bctconfig.yaml
			$ bctbackend ...

		For now, I will proceed with the default configuration,
		but this is strongly discouraged as it may lead to unexpected behavior.
		`,
		absolutePathOfConfigurationFile))
}
