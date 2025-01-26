package algorithms

// ContainsDuplicate returns the index of the first duplicate element in the given slice.
// If no duplicates are found, -1 is returned.
func ContainsDuplicate[T comparable](values []T) int {
	table := make(map[T]bool)

	for index := range values {
		value := &values[index]

		if table[*value] {
			return index
		}

		table[*value] = true
	}

	return -1
}
