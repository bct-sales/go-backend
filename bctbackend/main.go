package main

import (
	"bctbackend/commands"

	_ "modernc.org/sqlite"
)

func main() {
	commands.Execute()
}
