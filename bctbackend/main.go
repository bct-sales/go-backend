package main

import (
	"bctbackend/cli"
	"fmt"
	"os"
)

func main() {
	if err := cli.ProcessCommandLineArguments(os.Args); err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}
