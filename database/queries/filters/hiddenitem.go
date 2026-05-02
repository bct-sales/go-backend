package filters

import (
	"bctbackend/database/meta"
	"fmt"

	"github.com/Masterminds/squirrel"
)

type HiddenItem struct {
	filter
}

func NewHiddenItemFilter(query *WhereClauses) HiddenItem {
	return HiddenItem{
		filter: filter{
			storeClause: func(clause squirrel.Sqlizer) { query.addWhereClause(clause) },
		},
	}
}

func (filter *HiddenItem) WithHidden(isHidden bool) {
	columnName := fmt.Sprintf("%s.%s", meta.Item.Table, meta.Item.Hidden)
	clause := squirrel.Eq{columnName: isHidden}

	filter.storeClause(clause)
}
