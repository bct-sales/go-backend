package filters

type HiddenItem struct {
	hidden *bool
}

func (filter *HiddenItem) WithHidden(isHidden bool) {
	filter.hidden = &isHidden
}

// IsHidden returns the value set by WithHidden, or nil if it was never set.
func (filter *HiddenItem) IsHidden() *bool {
	return filter.hidden
}
