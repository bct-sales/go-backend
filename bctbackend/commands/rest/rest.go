package rest

import (
	"bctbackend/commands/common"
	"bctbackend/rest"
	"database/sql"
	"fmt"

	"github.com/spf13/cobra"
)

type RestCommand struct {
	common.Command
}

func NewRestCommand() *cobra.Command {
	var command *RestCommand

	command = &RestCommand{
		Command: common.Command{
			CobraCommand: &cobra.Command{
				Use:   "server",
				Short: "Start REST server",
				Long:  `This command starts the REST server for the BCT application.`,
				RunE: func(cmd *cobra.Command, args []string) error {
					return command.execute()
				},
			},
		},
	}

	return command.AsCobraCommand()
}

func (c *RestCommand) execute() error {
	return c.WithOpenedDatabase(func(db *sql.DB) error {
		if err := rest.StartRestService(db); err != nil {
			c.PrintErrorf("Failed to start REST service\n")
			return fmt.Errorf("failed to start REST service: %w", err)
		}

		return nil
	})
}
