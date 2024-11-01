package main

import (
	cli "bctbackend/cli"
	"os"
)

func main() {
	cli.ProcessCommandLineArguments(os.Args)
}
