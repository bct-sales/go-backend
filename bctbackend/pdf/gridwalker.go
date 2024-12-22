package pdf

type GridWalker struct {
	ColumnCount   int
	RowCount      int
	CurrentColumn int
	CurrentRow    int
}

func NewGridWalker(columnCount int, rowCount int) *GridWalker {
	return &GridWalker{
		ColumnCount:   columnCount,
		RowCount:      rowCount,
		CurrentColumn: 0,
		CurrentRow:    0,
	}
}

// Next moves the walker to the next cell in the grid.
func (gw *GridWalker) Next() {
	gw.CurrentColumn++

	if gw.CurrentColumn == gw.ColumnCount {
		gw.CurrentColumn = 0
		gw.CurrentRow++

		if gw.CurrentRow == gw.RowCount {
			gw.CurrentRow = 0
			return
		}
	}
}

func (gw *GridWalker) IsAtStart() bool {
	return gw.CurrentColumn == 0 && gw.CurrentRow == 0
}