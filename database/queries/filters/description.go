package filters

import (
	"bctbackend/database/meta"
	"fmt"

	"github.com/Masterminds/squirrel"
)

type Description struct {
	filter
}

func NewDescriptionFilter(query *WhereClauses) Category {
	return Category{
		filter: filter{
			storeClause: func(clause squirrel.Sqlizer) { query.addWhereClause(clause) },
		},
	}
}

func (filter *Category) WithDescriptionPattern(pattern string) {
	columnName := fmt.Sprintf("%s.%s", meta.Item.Table, meta.Item.Description)
	clause := squirrel.Like{columnName: pattern}

	filter.storeClause(clause)
}
