package filters

import "github.com/Masterminds/squirrel"

type WhereClauses struct {
	Clauses []squirrel.Sqlizer
}

func (query *WhereClauses) addWhereClause(clause squirrel.Sqlizer) {
	query.Clauses = append(query.Clauses, clause)
}

func NewWhereClauses() *WhereClauses {
	result := WhereClauses{Clauses: nil}

	return &result
}
