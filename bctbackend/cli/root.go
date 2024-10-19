package cli

import (
	"bctbackend/clilib"
	"fmt"
)

type RootCommand struct {
	Verbosity int
}

type ListCommand struct {
	RootCommand
}

type ListUsersCommand struct {
	ListCommand
}

type ListItemsCommand struct {
	ListCommand
}

func (command ListUsersCommand) Execute() {
	fmt.Println("Listing users")
}

func (command ListItemsCommand) Execute() {
	fmt.Println("Listing items")
}

func ParseCommandLineArguments(arguments []string) {
	parser := clilib.NewParser[RootCommand]()

	parser.Flag("-")
}
