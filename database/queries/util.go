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

type RowSelection struct {
	Limit  *uint64
	Offset *uint64
}

func (rowSelection *RowSelection) SQL() string {
	// If neither field has been set, no extra SQL is necessary
	if rowSelection.Limit == nil && rowSelection.Offset == nil {
		return ""
	}

	var offset uint64
	var limit uint64

	if rowSelection.Limit == nil {
		limit = 1000000
	} else {
		limit = *rowSelection.Limit
	}

	if rowSelection.Offset == nil {
		offset = 0
	} else {
		offset = *rowSelection.Offset
	}

	return fmt.Sprintf("LIMIT %d OFFSET %d", limit, offset)
}

func AllRows() *RowSelection {
	return &RowSelection{Limit: nil, Offset: nil}
}

func NewRowSelection(offset uint64, limit uint64) *RowSelection {
	return &RowSelection{Limit: &limit, Offset: &offset}
}

type Order int

const (
	OrderChronological Order = iota
	OrderAntiChronological
)
