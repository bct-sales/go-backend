package algorithms

import (
	"errors"
	"fmt"
	"os"
)

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

func MapError[T any, U any](values []T, f func(T) (U, error)) ([]U, error) {
	result := make([]U, len(values))

	for index, value := range values {
		transformedValue, err := f(value)
		if err != nil {
			return nil, fmt.Errorf("error when processing item with index %d: %w", index, err)
		}

		result[index] = transformedValue
	}

	return result, nil
}

func MapOptional[T any, U any](value *T, f func(T) U) *U {
	if value == nil {
		return nil
	}

	result := f(*value)

	return &result
}

func RepeatWithError(count int, function func() error) error {
	for range count {
		if err := function(); err != nil {
			return err
		}
	}

	return nil
}

func Repeat(count int, function func()) {
	for range count {
		function()
	}
}

func RepeatCollect[T any](count int, function func() T) []T {
	result := make([]T, count)

	for i := 0; i < count; i++ {
		result[i] = function()
	}

	return result
}

func Filter[T any](values []T, predicate func(T) bool) []T {
	result := make([]T, 0)

	for _, value := range values {
		if predicate(value) {
			result = append(result, value)
		}
	}

	return result
}

func RemoveDuplicates[T comparable](values []T) []T {
	set := NewSet(values...)
	return set.ToSlice()
}

func Range(start, end int) []int {
	size := max(end-start, 0)
	result := make([]int, size)

	for i := start; i != size; i++ {
		result[i] = start + i
	}

	return result
}

func FileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}

	return false, fmt.Errorf("failed to determine if file %s exists: %w", path, err)
}
