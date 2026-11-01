package queries

type RowRange struct {
	Limit  *uint64
	Offset *uint64
}

// LimitOffset returns the limit and offset to use for LIMIT/OFFSET query
// parameters. When neither limit nor offset is set, the limit is -1, which
// SQLite treats as "no upper bound", matching the old behavior of emitting no
// LIMIT clause at all. An offset without a limit keeps an implicit row cap
// (1000000), as the old generated SQL did.
func (rowRange *RowRange) LimitOffset() (limit int64, offset int64) {
	if rowRange.Limit == nil {
		if rowRange.Offset == nil {
			limit = -1
		} else {
			limit = 1000000
		}
	} else {
		limit = int64(*rowRange.Limit)
	}

	if rowRange.Offset == nil {
		offset = 0
	} else {
		offset = int64(*rowRange.Offset)
	}

	return limit, offset
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
