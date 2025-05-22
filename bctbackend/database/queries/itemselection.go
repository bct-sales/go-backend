package queries

import (
	"os"

	"golang.org/x/exp/slog"
)

type ItemSelection int

const (
	AllItems ItemSelection = iota
	OnlyVisibleItems
	OnlyHiddenItems
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

func ItemSelectionFromBool(onlyVisible bool) ItemSelection {
	if onlyVisible {
		return OnlyVisibleItems
	}
	return AllItems
}
