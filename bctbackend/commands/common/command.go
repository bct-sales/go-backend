package common

import (
	"fmt"

	"github.com/spf13/cobra"
)

type Command struct {
	CobraCommand *cobra.Command
}

func (c *Command) PrintError(formatString string, args ...any) {
	fmt.Fprintf(c.CobraCommand.ErrOrStderr(), formatString, args...)
}

func (c *Command) Print(formatString string, args ...any) {
	fmt.Fprintf(c.CobraCommand.OutOrStdout(), formatString, args...)
}
