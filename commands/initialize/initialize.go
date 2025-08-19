package initialize

import (
	"bctbackend/algorithms"
	"bctbackend/commands/common"
	"bctbackend/database"
	dberr "bctbackend/database/errors"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"
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

func (c *InitializeCommand) execute() error {
	if err := c.generateConfigurationFile(); err != nil {
		return err
	}

	if err := c.downloadHTMLFile(); err != nil {
		return err
	}

	if err := c.downloadFontFile(); err != nil {
		return err
	}

	if err := c.createDatabaseFile(); err != nil {
		return err
	}

	return nil
}

func (c *InitializeCommand) generateConfigurationFile() error {
	fileExists, err := algorithms.FileExists("bctconfig.yaml")
	if err != nil {
		c.PrintErrorf("Failed to check if configuration file exists: %v\n", err)
		return err
	}
	if fileExists {
		c.Printf("Configuration file already exists; I will not overwrite it\n")
		return nil
	}

	contents := heredoc.Doc(`
		database: "bct.db"
		labelGeneration:
		  barcode:
		    width: 150
		    height: 25
		  font:
		    directory: "."
		    filename: "noto.ttf"
		    family: "Noto"
		server:
		  port: 8000
		  html: index.html
		  pruneExpiredSessionsInterval: 3600
		  cookieDomain: "bct-sales.duckdns.org"
		  debug: false
		  swagger: false
		log:
		  file: "bct.log"
		  maxSizeMegabytes: 10
		  maxBackups: 3
		  maxAgeDays: 28
		  compression: false
	`)

	if err := os.WriteFile("bctconfig.yaml", []byte(contents), 0644); err != nil {
		c.PrintErrorf("Failed to create configuration file: %v\n", err)
		return err
	}

	c.Printf("Configuration file created successfully\n")
	return nil
}

func (c *InitializeCommand) downloadHTMLFile() (r_err error) {
	filename := "index.html"

	fileExists, err := algorithms.FileExists(filename)
	if err != nil {
		c.PrintErrorf("Failed to check if %s exists: %v\n", filename, err)
		return err
	}
	if fileExists {
		c.Printf("File %s already exists; I will not overwrite it\n", filename)
		return nil
	}

	out, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer func() {
		if err := out.Close(); err != nil {
			c.PrintErrorf("Failed to close file %s\n", filename)
			r_err = errors.Join(r_err, err)
		}
	}()

	url := "https://github.com/bct-sales/go-frontend/releases/latest/download/index.html"
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			c.PrintErrorf("Failed to close response body\n")
			r_err = errors.Join(r_err, err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		c.PrintErrorf("Failed to download file from %s: %s\n", url, resp.Status)
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	if _, err := io.Copy(out, resp.Body); err != nil {
		return err
	}

	c.Printf("HTML file downloaded successfully to %s\n", filename)

	return nil
}

func (c *InitializeCommand) downloadFontFile() (r_err error) {
	filename := "noto.ttf"

	fileExists, err := algorithms.FileExists(filename)
	if err != nil {
		c.PrintErrorf("Failed to check if %s exists: %v\n", filename, err)
		return err
	}
	if fileExists {
		c.Printf("File %s already exists; I will not overwrite it.\n", filename)
		return nil
	}

	out, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer func() {
		if err := out.Close(); err != nil {
			c.PrintErrorf("Failed to close file %s\n", filename)
			r_err = errors.Join(r_err, err)
		}
	}()

	url := "https://github.com/googlefonts/noto-fonts/raw/main/hinted/ttf/NotoSans/NotoSans-Regular.ttf"
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			c.PrintErrorf("Failed to close response body\n")
			r_err = errors.Join(r_err, err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		c.PrintErrorf("Failed to download file from %s: %s\n", url, resp.Status)
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	if _, err := io.Copy(out, resp.Body); err != nil {
		return err
	}

	c.Printf("Font file downloaded successfully to %s\n", filename)

	return nil
}

func (c *InitializeCommand) createDatabaseFile() (r_err error) {
	databasePath := "bct.db"
	db, err := database.CreateDatabase(databasePath)

	if err != nil {
		if errors.Is(err, dberr.ErrDatabaseAlreadyExists) {
			c.Printf("Database file %s already exists; I will not overwrite it.\n", databasePath)

			return err
		}

		c.PrintErrorf("Failed to create database file: %v\n", err)
		return err
	}

	defer func() {
		if err := db.Close(); err != nil {
			c.PrintErrorf("Failed to close database %s\n", databasePath)
			r_err = errors.Join(r_err, err)
		}
	}()

	c.Printf("Database file successfully created at %s\n", databasePath)

	if err := database.InitializeDatabase(db); err != nil {
		c.PrintErrorf("Failed to initialize database: %v\n", err)
		return err
	}

	c.Printf("Database initialized successfully\n")

	return nil
}
