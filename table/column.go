package table

import (
	"github.com/depado/gorich/console"
	"github.com/depado/gorich/style"
	"github.com/depado/gorich/table/align"
)

// Column defines a column within a Table.
type Column struct {
	Header      string
	Footer      string
	HeaderStyle *style.Style
	FooterStyle *style.Style
	Style       *style.Style
	Justify     console.Justify
	Vertical    align.VerticalMethod
	Overflow    console.Overflow
	Width       int // 0 = auto
	MinWidth    int
	MaxWidth    int
	Ratio       int // 0 = not flexible
	NoWrap      bool

	index int
	cells []console.Renderable
}

// newColumn creates a column with defaults.
func newColumn(header string, index int) *Column {
	return &Column{
		Header:   header,
		Justify:  console.JustifyLeft,
		Vertical: align.Top,
		Overflow: console.OverflowEllipsis,
		index:    index,
	}
}

// flexible returns true if this column has a ratio set.
func (c *Column) flexible() bool {
	return c.Ratio > 0
}

// ColumnOption configures a Column.
type ColumnOption func(*Column)

// WithColumnStyle sets the default style for cells in this column.
func WithColumnStyle(s string) ColumnOption {
	return func(c *Column) {
		st := style.Parse(s)
		c.Style = &st
	}
}

// WithColumnHeaderStyle sets the header style for this column.
func WithColumnHeaderStyle(s string) ColumnOption {
	return func(c *Column) {
		st := style.Parse(s)
		c.HeaderStyle = &st
	}
}

// WithColumnFooterStyle sets the footer style for this column.
func WithColumnFooterStyle(s string) ColumnOption {
	return func(c *Column) {
		st := style.Parse(s)
		c.FooterStyle = &st
	}
}

// WithColumnJustify sets the text justification for this column.
func WithColumnJustify(j console.Justify) ColumnOption {
	return func(c *Column) { c.Justify = j }
}

// WithColumnVertical sets the vertical alignment for this column.
func WithColumnVertical(v align.VerticalMethod) ColumnOption {
	return func(c *Column) { c.Vertical = v }
}

// WithColumnWidth sets a fixed width for this column.
func WithColumnWidth(w int) ColumnOption {
	return func(c *Column) { c.Width = w }
}

// WithColumnMinWidth sets a minimum width for this column.
func WithColumnMinWidth(w int) ColumnOption {
	return func(c *Column) { c.MinWidth = w }
}

// WithColumnMaxWidth sets a maximum width for this column.
func WithColumnMaxWidth(w int) ColumnOption {
	return func(c *Column) { c.MaxWidth = w }
}

// WithColumnRatio sets a ratio for flexible width distribution.
func WithColumnRatio(r int) ColumnOption {
	return func(c *Column) { c.Ratio = r }
}

// WithColumnNoWrap disables text wrapping for this column.
func WithColumnNoWrap() ColumnOption {
	return func(c *Column) { c.NoWrap = true }
}

// WithColumnOverflow sets the overflow behavior for this column.
func WithColumnOverflow(o console.Overflow) ColumnOption {
	return func(c *Column) { c.Overflow = o }
}
