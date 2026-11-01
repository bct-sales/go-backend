//go:build test

package queries

import (
	"bctbackend/database/queries"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRowRangeLimitOffset(t *testing.T) {
	t.Run("All rows yields an unbounded limit", func(t *testing.T) {
		// A query with neither limit nor offset set used to have no LIMIT
		// clause at all, i.e., it was unbounded. The sentinel for "no upper
		// bound" is a negative limit, which SQLite treats as unlimited.
		limit, offset := queries.AllRows().LimitOffset()

		require.Equal(t, int64(-1), limit)
		require.Equal(t, int64(0), offset)
	})

	t.Run("Explicit limit and offset are passed through", func(t *testing.T) {
		limit, offset := queries.NewRowRange(10, 5).LimitOffset()

		require.Equal(t, int64(5), limit)
		require.Equal(t, int64(10), offset)
	})

	t.Run("Offset without limit keeps the implicit row cap", func(t *testing.T) {
		// An offset without a limit always had an implicit cap (previously
		// LIMIT 1000000 in the generated SQL).
		rowRange := queries.RowRange{Offset: newUint64(10)}
		limit, offset := rowRange.LimitOffset()

		require.Equal(t, int64(1000000), limit)
		require.Equal(t, int64(10), offset)
	})
}

func newUint64(value uint64) *uint64 {
	return &value
}
