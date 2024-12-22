package pdf

type LayoutSettings struct {
	PaperWidth        float64
	PaperHeight       float64
	PaperTopMargin    float64
	PaperBottomMargin float64
	PaperLeftMargin   float64
	PaperRightMargin  float64
	Columns           int
	Rows              int
	LabelMargin       float64
	LabelPadding      float64
	FontSize          float64
}

type ValidatedLayoutSettings struct {
	paperWidth        float64
	paperHeight       float64
	paperTopMargin    float64
	paperBottomMargin float64
	paperLeftMargin   float64
	paperRightMargin  float64
	columns           int
	rows              int
	labelMargin       float64
	labelPadding      float64
	fontSize          float64
}

func NewLayoutSettings() *LayoutSettings {
	return &LayoutSettings{}
}

func (ls *LayoutSettings) SetA4PaperSize() *LayoutSettings {
	ls.PaperWidth = 210.0
	ls.PaperHeight = 297.0

	return ls
}

func (ls *LayoutSettings) SetPaperMargins(margin float64) *LayoutSettings {
	ls.PaperTopMargin = margin
	ls.PaperBottomMargin = margin
	ls.PaperLeftMargin = margin
	ls.PaperRightMargin = margin

	return ls
}

func (ls *LayoutSettings) SetGridSize(columns int, rows int) *LayoutSettings {
	ls.Columns = columns
	ls.Rows = rows

	return ls
}

func (ls *LayoutSettings) SetLabelMargin(margin float64) *LayoutSettings {
	ls.LabelMargin = margin

	return ls
}

func (ls *LayoutSettings) SetLabelPadding(padding float64) *LayoutSettings {
	ls.LabelPadding = padding

	return ls
}

func (ls *LayoutSettings) SetFontSize(size float64) *LayoutSettings {
	ls.FontSize = size

	return ls
}

func (ls *LayoutSettings) Validate() *ValidatedLayoutSettings {
	if ls.PaperWidth <= 0 {
		return nil
	}

	if ls.PaperHeight <= 0 {
		return nil
	}

	if ls.PaperTopMargin < 0 {
		return nil
	}

	if ls.PaperBottomMargin < 0 {
		return nil
	}

	if ls.PaperLeftMargin < 0 {
		return nil
	}

	if ls.PaperRightMargin < 0 {
		return nil
	}

	if ls.Columns <= 0 {
		return nil
	}

	if ls.Rows <= 0 {
		return nil
	}

	if ls.LabelMargin < 0 {
		return nil
	}

	if ls.LabelPadding < 0 {
		return nil
	}

	if ls.FontSize <= 0 {
		return nil
	}

	return &ValidatedLayoutSettings{
		paperWidth:        ls.PaperWidth,
		paperHeight:       ls.PaperHeight,
		paperTopMargin:    ls.PaperTopMargin,
		paperBottomMargin: ls.PaperBottomMargin,
		paperLeftMargin:   ls.PaperLeftMargin,
		paperRightMargin:  ls.PaperRightMargin,
		columns:           ls.Columns,
		rows:              ls.Rows,
		labelMargin:       ls.LabelMargin,
		labelPadding:      ls.LabelPadding,
		fontSize:          ls.FontSize,
	}
}

func (ls *ValidatedLayoutSettings) GetColumnWidth() float64 {
	return (ls.paperWidth - ls.paperLeftMargin - ls.paperRightMargin) / float64(ls.columns)
}

func (ls *ValidatedLayoutSettings) GetRowHeight() float64 {
	return (ls.paperHeight - ls.paperTopMargin - ls.paperBottomMargin) / float64(ls.rows)
}

func (ls *ValidatedLayoutSettings) GetRectangle(column int, row int) *Rectangle {
	return (&Rectangle{
		Left:   ls.paperLeftMargin + float64(column)*ls.GetColumnWidth(),
		Top:    ls.paperTopMargin + float64(row)*ls.GetRowHeight(),
		Width:  ls.GetColumnWidth(),
		Height: ls.GetRowHeight(),
	}).ShrinkUniformly(ls.labelMargin)
}

func (ls *ValidatedLayoutSettings) GetColumns() int {
	return ls.columns
}

func (ls *ValidatedLayoutSettings) GetRows() int {
	return ls.rows
}

func IsA4Size(layout *ValidatedLayoutSettings) bool {
	return layout.paperWidth == 210.0 && layout.paperHeight == 297.0
}
