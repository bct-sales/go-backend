package queries

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
		panic("Invalid hidden strategy: " + string(itemSelection))
	}
}

func ItemSelectionFromBool(onlyVisible bool) ItemSelection {
	if onlyVisible {
		return OnlyVisibleItems
	}
	return AllItems
}
