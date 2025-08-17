package pdf

import (
	"errors"
	"fmt"
)

type LayoutSettings struct {
	paperWidth   float64
	paperHeight  float64
	paperMargins Insets
	columns      int
	rows         int
	labelMargins Insets
	labelPadding Insets
	fontSize     float64
}

type layoutSettingsOption func(*LayoutSettings)

func NewLayoutSettings(options ...layoutSettingsOption) (*LayoutSettings, error) {
	var layoutSettings LayoutSettings

	for _, option := range options {
		option(&layoutSettings)
	}

	if err := Validate(&layoutSettings); err != nil {
		return nil, err
	}

	return &layoutSettings, nil
}

func WithPaperSize(width float64, height float64) layoutSettingsOption {
	return func(ls *LayoutSettings) {
		ls.paperWidth = width
		ls.paperHeight = height
	}
}

func WithA4PaperSize() layoutSettingsOption {
	return WithPaperSize(210.0, 297.0)
}

func WithPaperMargins(top float64, right float64, bottom float64, left float64) layoutSettingsOption {
	return func(ls *LayoutSettings) {
		ls.paperMargins = Insets{
			Top:    top,
			Right:  right,
			Bottom: bottom,
			Left:   left,
		}
	}
}

func WithUniformPaperMargin(margin float64) layoutSettingsOption {
	return WithPaperMargins(margin, margin, margin, margin)
}

func WithGridSize(columns int, rows int) layoutSettingsOption {
	return func(ls *LayoutSettings) {
		ls.columns = columns
		ls.rows = rows
	}
}

func WithLabelMargins(top float64, right float64, bottom float64, left float64) layoutSettingsOption {
	return func(ls *LayoutSettings) {
		ls.labelMargins = Insets{
			Top:    top,
			Right:  right,
			Bottom: bottom,
			Left:   left,
		}
	}
}

func WithUniformLabelMargin(margin float64) layoutSettingsOption {
	return WithLabelMargins(margin, margin, margin, margin)
}

func WithLabelPadding(top float64, right float64, bottom float64, left float64) layoutSettingsOption {
	return func(ls *LayoutSettings) {
		ls.labelPadding = Insets{
			Top:    top,
			Right:  right,
			Bottom: bottom,
			Left:   left,
		}
	}
}

func WithUniformLabelPadding(padding float64) layoutSettingsOption {
	return WithLabelPadding(padding, padding, padding, padding)
}

func WithFontSize(size float64) layoutSettingsOption {
	return func(ls *LayoutSettings) {
		ls.fontSize = size
	}
}

func Validate(layoutSettings *LayoutSettings) error {
	var errs []error

	errorIf := func(condition bool, message string) {
		if condition {
			errs = append(errs, fmt.Errorf("%s", message))
		}
	}

	errorIf(layoutSettings.paperWidth <= 0, "paper width must be greater than 0")
	errorIf(layoutSettings.paperHeight <= 0, "paper height must be greater than 0")
	errorIf(layoutSettings.paperMargins.Top < 0, "paper top margin must be greater than or equal to 0")
	errorIf(layoutSettings.paperMargins.Bottom < 0, "paper bottom margin must be greater than or equal to 0")
	errorIf(layoutSettings.paperMargins.Left < 0, "paper left margin must be greater than or equal to 0")
	errorIf(layoutSettings.paperMargins.Right < 0, "paper right margin must be greater than or equal to 0")
	errorIf(layoutSettings.columns <= 0, "number of columns must be greater than 0")
	errorIf(layoutSettings.rows <= 0, "number of rows must be greater than 0")
	errorIf(layoutSettings.labelMargins.Top < 0, "top label margin must be greater than or equal to 0")
	errorIf(layoutSettings.labelMargins.Bottom < 0, "bottom label margin must be greater than or equal to 0")
	errorIf(layoutSettings.labelMargins.Left < 0, "left label margin must be greater than or equal to 0")
	errorIf(layoutSettings.labelMargins.Right < 0, "right label margin must be greater than or equal to 0")
	errorIf(layoutSettings.labelPadding.Top < 0, "top label padding must be greater than or equal to 0")
	errorIf(layoutSettings.labelPadding.Bottom < 0, "bottom label padding must be greater than or equal to 0")
	errorIf(layoutSettings.labelPadding.Left < 0, "left label padding must be greater than or equal to 0")
	errorIf(layoutSettings.labelPadding.Right < 0, "right label padding must be greater than or equal to 0")
	errorIf(layoutSettings.fontSize <= 0, "font size must be greater than 0")

	return errors.Join(errs...)
}

func (ls *LayoutSettings) GetColumnWidth() float64 {
	return (ls.paperWidth - ls.paperMargins.Left - ls.paperMargins.Right) / float64(ls.columns)
}

func (ls *LayoutSettings) GetRowHeight() float64 {
	return (ls.paperHeight - ls.paperMargins.Top - ls.paperMargins.Bottom) / float64(ls.rows)
}

func (ls *LayoutSettings) GetRectangle(column int, row int) *Rectangle {
	return (&Rectangle{
		Left:   ls.paperMargins.Left + float64(column)*ls.GetColumnWidth(),
		Top:    ls.paperMargins.Top + float64(row)*ls.GetRowHeight(),
		Width:  ls.GetColumnWidth(),
		Height: ls.GetRowHeight(),
	}).Shrink(ls.labelMargins)
}

func (ls *LayoutSettings) GetColumns() int {
	return ls.columns
}

func (ls *LayoutSettings) GetRows() int {
	return ls.rows
}

func IsA4Size(layout *LayoutSettings) bool {
	return layout.paperWidth == 210.0 && layout.paperHeight == 297.0
}
