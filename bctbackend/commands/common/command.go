package common

import (
	"fmt"

	"github.com/spf13/cobra"
)

type Command struct {
	cobraCommand *cobra.Command
}

func (c *Command) PrintError(formatString string, args ...any) {
	fmt.Fprintf(c.cobraCommand.ErrOrStderr(), formatString, args...)
}

func (c *Command) Print(formatString string, args ...any) {
	fmt.Fprintf(c.cobraCommand.OutOrStdout(), formatString, args...)
}
