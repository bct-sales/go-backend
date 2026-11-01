package filters

type LargeItem struct {
	large *bool
}

func (filter *LargeItem) WithLarge(isLargeItem bool) {
	filter.large = &isLargeItem
}

// IsLarge returns the value set by WithLarge, or nil if it was never set.
func (filter *LargeItem) IsLarge() *bool {
	return filter.large
}
