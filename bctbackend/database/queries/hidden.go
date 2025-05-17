package queries

import (
	"os"

	"golang.org/x/exp/slog"
)

const (
	AllItems         = 1
	OnlyVisibleItems = 2
	OnlyHiddenItems  = 3
)

func ItemsTableFor(hiddenStrategy int) string {
	switch hiddenStrategy {
	case AllItems:
		return "items"
	case OnlyVisibleItems:
		return "visible_items"
	case OnlyHiddenItems:
		return "hidden_items"
	default:
		slog.Error("Invalid hidden strategy", "hiddenStrategy", hiddenStrategy)
		os.Exit(1)
		return ""
	}
}
