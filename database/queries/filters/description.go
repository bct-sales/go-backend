package filters

import (
	"bctbackend/database/meta"
	"bctbackend/database/models"
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

func (filter *Category) WithDescriptionPattern(category models.ID) {
	columnName := fmt.Sprintf("%s.%s", meta.Item.Table, meta.Item.ItemCategoryID)
	clause := squirrel.Eq{columnName: category}

	filter.storeClause(clause)
}
