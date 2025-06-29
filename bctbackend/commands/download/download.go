package download

import (
	"bctbackend/commands/common"
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
			},
		},
	}

	return command.AsCobraCommand()
}

func (c *DownloadCommand) execute() error {
	out, err := os.Create("index.html")
	if err != nil {
		return err
	}
	defer out.Close()

	url := "https://github.com/bct-sales/go-frontend/releases/latest/download/index.html"
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.PrintErrorf("Failed to download file from %s: %s\n", url, resp.Status)
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	c.Printf("HTML file downloaded successfully to index.html\n")

	return nil
}
