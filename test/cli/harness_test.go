//go:build test

package cli_tests

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"bctbackend/database"

	// The sqlite driver is needed to open database files directly in-process.
	_ "modernc.org/sqlite"

	"github.com/stretchr/testify/require"
)

var binaryPath string

// commandTimeout bounds how long a single CLI invocation may run.
var commandTimeout = 60 * time.Second

// TestHarnessCommandTimeout verifies that runCommand kills CLI invocations
// that hang instead of blocking the whole test suite.
func TestHarnessCommandTimeout(t *testing.T) {
	originalBinaryPath := binaryPath
	originalTimeout := commandTimeout
	defer func() {
		binaryPath = originalBinaryPath
		commandTimeout = originalTimeout
	}()

	// A command that takes much longer than the timeout.
	if runtime.GOOS == "windows" {
		binaryPath = "ping"
		commandTimeout = 2 * time.Second

		result := runCommand(t, t.TempDir(), nil, []string{"-n", "10", "127.0.0.1"}...)

		require.Equal(t, -1, result.ExitCode)
		require.Contains(t, result.Stderr, "timed out")
	} else {
		binaryPath = "sleep"
		commandTimeout = 2 * time.Second

		result := runCommand(t, t.TempDir(), nil, []string{"30"}...)

		require.Equal(t, -1, result.ExitCode)
		require.Contains(t, result.Stderr, "timed out")
	}
}

// TestMain builds the bctbackend binary once, so that every test can run it
// as a subprocess.
func TestMain(m *testing.M) {
	directory, err := os.MkdirTemp("", "bct-cli-tests-binary")
	if err != nil {
		panic("failed to create temporary directory for binary: " + err.Error())
	}

	executable := "bctbackend"
	if runtime.GOOS == "windows" {
		executable += ".exe"
	}

	binaryPath = filepath.Join(directory, executable)

	build := exec.Command("go", "build", "-o", binaryPath, ".")
	build.Dir = filepath.Join(mustGetWorkingDirectory(), "..", "..")

	if output, err := build.CombinedOutput(); err != nil {
		panic("failed to build bctbackend binary: " + err.Error() + "\n" + string(output))
	}

	code := m.Run()

	os.RemoveAll(directory)
	os.Exit(code)
}

func mustGetWorkingDirectory() string {
	directory, err := os.Getwd()
	if err != nil {
		panic("failed to get working directory: " + err.Error())
	}

	return directory
}

// CommandResult holds the outcome of one CLI invocation.
type CommandResult struct {
	ExitCode int
	Stdout   string
	Stderr   string
}

// Workspace is a temporary directory in which CLI commands run.
// It contains an empty bctconfig.yaml so that configuration loading is a no-op.
type Workspace struct {
	Directory string
}

// NewWorkspace creates a fresh temporary workspace directory.
func NewWorkspace(t *testing.T) *Workspace {
	t.Helper()

	directory := t.TempDir()

	// The root command loads ./bctconfig.yaml from the working directory and
	// only logs a warning when it is absent; an empty file keeps the output clean.
	configPath := filepath.Join(directory, "bctconfig.yaml")
	require.NoError(t, os.WriteFile(configPath, []byte{}, 0o644))

	return &Workspace{Directory: directory}
}

// Path returns the path of the file with the given name inside the workspace.
func (w *Workspace) Path(name string) string {
	return filepath.Join(w.Directory, name)
}

// DatabasePath returns the default path of the workspace's database file.
func (w *Workspace) DatabasePath() string {
	return w.Path("bct.db")
}

// Run runs the bctbackend binary in this workspace.
func (w *Workspace) Run(t *testing.T, arguments ...string) CommandResult {
	t.Helper()

	return w.RunWithEnv(t, nil, arguments...)
}

// RunWithEnv runs the bctbackend binary in this workspace
// with additional environment variables ("KEY=value" strings).
func (w *Workspace) RunWithEnv(t *testing.T, environment []string, arguments ...string) CommandResult {
	t.Helper()

	return runCommand(t, w.Directory, environment, arguments...)
}

// runCommand runs the bctbackend binary in the given workspace directory
// and captures its exit code, stdout and stderr.
// A command that runs longer than commandTimeout is killed and reported
// with exit code -1 and a "timed out" message on stderr.
func runCommand(t *testing.T, workspace string, environment []string, arguments ...string) CommandResult {
	t.Helper()

	var stdout, stderr bytes.Buffer

	ctx, cancel := context.WithTimeout(context.Background(), commandTimeout)
	defer cancel()

	command := exec.CommandContext(ctx, binaryPath, arguments...)
	command.Dir = workspace
	command.Env = append(os.Environ(), environment...)
	command.Stdout = &stdout
	command.Stderr = &stderr

	err := command.Run()

	if ctx.Err() != nil {
		return CommandResult{
			ExitCode: -1,
			Stdout:   stdout.String(),
			Stderr:   fmt.Sprintf("command timed out after %s: %v\n%s", commandTimeout, arguments, stderr.String()),
		}
	}

	exitCode := 0
	if err != nil {
		var exitError *exec.ExitError
		require.ErrorAs(t, err, &exitError)
		exitCode = exitError.ExitCode()
	}

	return CommandResult{
		ExitCode: exitCode,
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
	}
}

// requireSuccess asserts that the command exited with code 0.
func requireSuccess(t *testing.T, result CommandResult) {
	t.Helper()
	require.Equal(t, 0, result.ExitCode, "command failed\nstdout:\n%s\nstderr:\n%s", result.Stdout, result.Stderr)
}

// requireFailure asserts that the command exited with a nonzero code.
func requireFailure(t *testing.T, result CommandResult) {
	t.Helper()
	require.NotEqual(t, 0, result.ExitCode, "command unexpectedly succeeded\nstdout:\n%s\nstderr:\n%s", result.Stdout, result.Stderr)
}

// SeedDatabase opens the given initialized database file, passes it to the
// seed function and closes it again.
func SeedDatabase(t *testing.T, databasePath string, seed func(db *sql.DB)) {
	t.Helper()

	database, err := database.OpenDatabase(databasePath)
	require.NoError(t, err)
	defer database.Close()

	seed(database)
}

// VerifyDatabase opens the given initialized database file, passes it to the
// verify function and closes it again. Intended for read-only checks.
func VerifyDatabase(t *testing.T, databasePath string, verify func(t *testing.T, db *sql.DB)) {
	t.Helper()

	SeedDatabase(t, databasePath, func(db *sql.DB) {
		verify(t, db)
	})
}
