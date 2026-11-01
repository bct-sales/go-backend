package queries

import (
	"database/sql"
	"fmt"
)

type ItemSelection int

const (
	AllItems ItemSelection = iota
	OnlyVisibleItems
	OnlyHiddenItems
)

// itemSelectionToNullBool maps an item selection to the nullable `hidden`
// filter used by generated queries: no filter for AllItems, otherwise an
// equality filter on the hidden column.
func itemSelectionToNullBool(itemSelection ItemSelection) sql.NullBool {
	switch itemSelection {
	case AllItems:
		return sql.NullBool{}
	case OnlyVisibleItems:
		return sql.NullBool{Bool: false, Valid: true}
	case OnlyHiddenItems:
		return sql.NullBool{Bool: true, Valid: true}
	default:
		panic(fmt.Sprintf("Invalid hidden strategy: %d", itemSelection))
	}
}

func ItemsTableFor(itemSelection ItemSelection) string {
	switch itemSelection {
	case AllItems:
		return "items"
	case OnlyVisibleItems:
		return "visible_items"
	case OnlyHiddenItems:
		return "hidden_items"
	default:
		panic(fmt.Sprintf("Invalid hidden strategy: %d", itemSelection))
	}
}

func ItemSelectionFromBool(onlyVisible bool) ItemSelection {
	if onlyVisible {
		return OnlyVisibleItems
	}
	return AllItems
}
