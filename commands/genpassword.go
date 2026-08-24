package commands

import (
	"bctbackend/algorithms"
	"bctbackend/commands/common"
	"bctbackend/commands/user"
	"bctbackend/database/queries"
	"database/sql"
	"errors"
	"fmt"
	"math/rand"

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

var ErrOutOfPasswords = errors.New("out of passwords")

func (c *generatePasswordCommand) execute() error {
	return c.WithOpenedDatabase(func(db *sql.DB) error {
		passwords, err := queries.CollectPasswords(db)
		if err != nil {
			return err
		}
		passwordSet := algorithms.NewSet(passwords...)

		indices := algorithms.Range(0, len(passwords))
		rand.Shuffle(len(indices), func(i, j int) {
			indices[i], indices[j] = indices[j], indices[i]
		})

		for _, index := range indices {
			candidatePassword := user.Passwords[index]

			if !passwordSet.Contains(candidatePassword) {
				fmt.Println(candidatePassword)
				return nil
			}
		}

		return ErrOutOfPasswords
	})
}
