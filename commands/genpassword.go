package commands

import (
	"bctbackend/commands/common"
	"database/sql"
	"fmt"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"
)

type generatePasswordCommand struct {
	common.Command
}

func NewGeneratePasswordCommand() *cobra.Command {
	var command *generatePasswordCommand

	command = &generatePasswordCommand{
		Command: common.Command{
			CobraCommand: &cobra.Command{
				Use:   "generate-password",
				Short: "Generates password",
				Long: heredoc.Doc(`
						Generates a password that has hitherto not been assigned to any user in the database.
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

func (c *generatePasswordCommand) execute() error {
	return c.WithOpenedDatabase(func(db *sql.DB) error {
		password, err := common.FindUnusedPassword(db)

		if err != nil {
			return err
		}

		fmt.Println(password)
		return nil
	})
}
