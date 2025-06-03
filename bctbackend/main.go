package main

import (
	"bctbackend/cli"
	"fmt"
	"os"

	_ "modernc.org/sqlite"
)

func main() {
	if err := cli.ProcessCommandLineArguments(os.Args); err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}
