package filters

import "github.com/Masterminds/squirrel"

type filter struct {
	storeClause func(squirrel.Sqlizer)
}

func AssignTo(target *squirrel.Sqlizer) func(squirrel.Sqlizer) {
	return func(clause squirrel.Sqlizer) {
		*target = clause
	}
}

func AppendTo(slice *[]squirrel.Sqlizer) func(squirrel.Sqlizer) {
	return func(clause squirrel.Sqlizer) {
		*slice = append(*slice, clause)
	}
}
