package download

import (
	"bctbackend/commands/common"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/spf13/cobra"
)

type DownloadCommand struct {
	common.Command
}

func NewDownloadCommand() *cobra.Command {
	var command *DownloadCommand

	command = &DownloadCommand{
		Command: common.Command{
			CobraCommand: &cobra.Command{
				Use:   "download",
				Short: "Download HTML",
				Long:  `This command fetches the latest version of the HTML file from GitHub.`,
				RunE: func(cmd *cobra.Command, args []string) error {
					return command.execute()
				},
				Args: cobra.NoArgs,
			},
		},
	}

	return command.AsCobraCommand()
}

func (c *DownloadCommand) execute() (r_err error) {
	filename := "index.html"
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

	c.Printf("HTML file downloaded successfully to index.html\n")

	return nil
}
