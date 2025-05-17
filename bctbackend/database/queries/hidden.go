package queries

import (
	"os"

	"golang.org/x/exp/slog"
)

type ItemSelection int

const (
	AllItems         ItemSelection = 1
	OnlyVisibleItems ItemSelection = 2
	OnlyHiddenItems  ItemSelection = 3
)

func ItemsTableFor(itemSelection ItemSelection) string {
	switch itemSelection {
	case AllItems:
		return "items"
	case OnlyVisibleItems:
		return "visible_items"
	case OnlyHiddenItems:
		return "hidden_items"
	default:
		slog.Error("Invalid hidden strategy", "hiddenStrategy", itemSelection)
		os.Exit(1)
		return ""
	}
}
