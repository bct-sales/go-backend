package main

import (
	cli "bctbackend/cli"
	"fmt"
	"os"
)

func main() {
	if err := cli.ProcessCommandLineArguments(os.Args); err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}
