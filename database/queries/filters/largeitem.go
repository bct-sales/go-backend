package filters

import (
	"bctbackend/database/meta"
	"fmt"

	"github.com/Masterminds/squirrel"
)

type LargeItem struct {
	filter
}

func NewLargeItemFilter(query *WhereClauses) LargeItem {
	return LargeItem{
		filter: filter{
			storeClause: func(clause squirrel.Sqlizer) { query.addWhereClause(clause) },
		},
	}
}

func (filter *LargeItem) WithLarge(isLargeItem bool) {
	columnName := fmt.Sprintf("%s.%s", meta.Item.Table, meta.Item.Large)
	clause := squirrel.Eq{columnName: isLargeItem}

	filter.storeClause(clause)
}
