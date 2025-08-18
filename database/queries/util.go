package queries

import (
	"fmt"
	"strconv"
	"strings"
)

func placeholderString(placeholderCount int) string {
	placeholders := make([]string, placeholderCount)
	for i := range placeholderCount {
		placeholders[i] = "?"
	}

	return strings.Join(placeholders, ", ")
}

type SQLOption interface {
	SQL() string
}

type RowSelection struct {
	Limit  *int
	Offset *int
}

func (p *RowSelection) SQL() string {
	clause := "LIMIT "

	if p.Limit != nil {
		clause += strconv.FormatInt(int64(*p.Limit), 10)
	} else {
		clause += "1000000"
	}

	if p.Offset != nil {
		clause += fmt.Sprintf(" OFFSET %d", *p.Offset)
	}

	return clause
}

func AllRows() *RowSelection {
	return &RowSelection{Limit: nil, Offset: nil}
}

func NewRowSelection(offset *int, limit *int) *RowSelection {
	return &RowSelection{Limit: limit, Offset: offset}
}
