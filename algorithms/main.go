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

// Map applies the given function f to each value and collects results in a slice.
// This version of Map expects f not to return error values.
func Map[T any, U any](values []T, f func(T) U) []U {
	result := make([]U, len(values))

	for index := range values {
		result[index] = f(values[index])
	}

	return result
}

// MapError applies the given function f to each value and collects results in a slice.
// It expects f to return an error value.
// As soon as one error is encountered, the processing stops and the error is returned.
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

// MapOptional applies the given function f to the value if it is not nil.
// If the value is nil, it returns nil.
func MapOptional[T any, U any](value *T, f func(T) U) *U {
	if value == nil {
		return nil
	}

	result := f(*value)

	return &result
}

// RepeatWithError calls the given function count times.
// If any invocation returns an error, the error is returned immediately.
func RepeatWithError(count int, function func() error) error {
	for range count {
		if err := function(); err != nil {
			return err
		}
	}

	return nil
}

// Repeat calls the given function count times.
func Repeat(count int, function func()) {
	for range count {
		function()
	}
}

// RepeatCollect calls the given function count times and collects the results in a slice.
func RepeatCollect[T any](count int, function func() T) []T {
	result := make([]T, count)

	for i := range count {
		result[i] = function()
	}

	return result
}

// Filter selects the values that satisfy the given predicate.
// The order of the values is preserved.
func Filter[T any](values []T, predicate func(T) bool) []T {
	result := make([]T, 0)

	for _, value := range values {
		if predicate(value) {
			result = append(result, value)
		}
	}

	return result
}

// RemoveDuplicates removes duplicate values from the slice.
// The order is not necessarily preserved.
func RemoveDuplicates[T comparable](values []T) []T {
	set := NewSet(values...)
	return set.ToSlice()
}

// Range returns a slice of integers from start to end (exclusive).
func Range(start, end int) []int {
	size := max(end-start, 0)
	result := make([]int, size)

	for i := start; i != size; i++ {
		result[i] = start + i
	}

	return result
}

// FileExists checks if a file exists at the given path.
// Returns true if the file exists, false otherwise.
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

// Any checks if any value in values satisfies the given predicate.
func Any[T any](values []T, predicate func(T) bool) bool {
	for _, value := range values {
		if predicate(value) {
			return true
		}
	}

	return false
}
