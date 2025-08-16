package initialize

import (
	"bctbackend/commands/common"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"
	viperlib "github.com/spf13/viper"
)

type InitializeCommand struct {
	common.Command
}

func NewInitializeCommand() *cobra.Command {
	var command *InitializeCommand

	command = &InitializeCommand{
		Command: common.Command{
			CobraCommand: &cobra.Command{
				Use:   "init",
				Short: "Creates configuration file",
				Long: heredoc.Doc(`
							This command creates a configuration file for the BCT application.
							It will create a file named 'bctconfig.yaml' in the current directory.
					   `),
				RunE: func(cmd *cobra.Command, args []string) error {
					return command.execute()
				},
				Args: cobra.NoArgs,
			},
		},
	}

	return command.AsCobraCommand()
}

func (command *InitializeCommand) execute() error {
	settingsCopy := command.copySettings()

	if err := settingsCopy.SafeWriteConfig(); err != nil {
		command.Printf("Failed to create configuration file: %v\n", err)
		return err
	}

	command.Printf("Configuration file created successfully\n")
	return nil
}

func (command *InitializeCommand) copySettings() *viperlib.Viper {
	viper := viperlib.New()

	viper.SetConfigName("bctconfig")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	for key, value := range viper.AllSettings() {
		// Skip the "config" key, it's a bit silly to have the config file reference itself
		if key != "config" {
			viper.Set(key, value)
		}
	}

	return viper
}
