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

func Map[T any, U any](values []T, f func(T) U) []U {
	result := make([]U, len(values))

	for index := range values {
		result[index] = f(values[index])
	}

	return result
}

func Repeat(count int, function func() error) error {
	for i := 0; i < count; i++ {
		if err := function(); err != nil {
			return err
		}
	}

	return nil
}
