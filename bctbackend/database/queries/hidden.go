package queries

import (
	"os"

	"golang.org/x/exp/slog"
)

const (
	IncludeHidden = 1
	ExcludeHidden = 2
	OnlyHidden    = 3
)

func ItemsTableFor(hiddenStrategy int) string {
	switch hiddenStrategy {
	case IncludeHidden:
		return "items"
	case ExcludeHidden:
		return "visible_items"
	case OnlyHidden:
		return "hidden_items"
	default:
		slog.Error("Invalid hidden strategy", "hiddenStrategy", hiddenStrategy)
		os.Exit(1)
		return ""
	}
}
