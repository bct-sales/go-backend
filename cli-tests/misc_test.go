//go:build test

package cli_tests

import (
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMiscCommands(t *testing.T) {
	t.Run("Version prints the version", func(t *testing.T) {
		workspace := NewWorkspace(t)

		result := workspace.Run(t, "version")

		requireSuccess(t, result)
		require.Regexp(t, regexp.MustCompile(`^\d+\.\d+\.\d+\n$`), result.Stdout)
	})

	t.Run("Version rejects arguments", func(t *testing.T) {
		workspace := NewWorkspace(t)

		result := workspace.Run(t, "version", "extra")

		requireFailure(t, result)
	})

	t.Run("Generate-password prints a single line", func(t *testing.T) {
		run, databasePath := setupDatabaseTest(t)
		requireSuccess(t, run("user", "add", "--id", "1", "--role", "admin", "--password", "pw1", "--db", databasePath))

		result := run("generate-password", "--db", databasePath)

		requireSuccess(t, result)
		require.Regexp(t, regexp.MustCompile(`^\S+\n$`), result.Stdout)
	})

	t.Run("Generate-password avoids used passwords", func(t *testing.T) {
		run, databasePath := setupDatabaseTest(t)
		requireSuccess(t, run("user", "add", "--id", "1", "--role", "admin", "--password", "pw1", "--db", databasePath))

		first := run("generate-password", "--db", databasePath)
		requireSuccess(t, first)
		password := strings.TrimSpace(first.Stdout)

		requireSuccess(t, run("user", "add", "--id", "2", "--role", "admin", "--password", password, "--db", databasePath))

		second := run("generate-password", "--db", databasePath)

		requireSuccess(t, second)
		require.NotEqual(t, first.Stdout, second.Stdout)
	})
}
