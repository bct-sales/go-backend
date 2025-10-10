package queries

import (
	"fmt"
	"strings"
)

func placeholderString(placeholderCount int) string {
	placeholders := make([]string, placeholderCount)
	for i := range placeholderCount {
		placeholders[i] = "?"
	}

	return strings.Join(placeholders, ", ")
}

type RowRange struct {
	Limit  *uint64
	Offset *uint64
}

func (rowRange *RowRange) SQL() string {
	// If neither field has been set, no extra SQL is necessary
	if rowRange.Limit == nil && rowRange.Offset == nil {
		return ""
	}

	var offset uint64
	var limit uint64

	if rowRange.Limit == nil {
		limit = 1000000
	} else {
		limit = *rowRange.Limit
	}

	if rowRange.Offset == nil {
		offset = 0
	} else {
		offset = *rowRange.Offset
	}

	return fmt.Sprintf("LIMIT %d OFFSET %d", limit, offset)
}

func AllRows() *RowRange {
	return &RowRange{Limit: nil, Offset: nil}
}

func NewRowRange(offset uint64, limit uint64) *RowRange {
	return &RowRange{Limit: &limit, Offset: &offset}
}

type Order int

const (
	OrderChronological Order = iota
	OrderAntiChronological
)
