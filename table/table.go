package table

import (
	"fmt"

	"github.com/depado/gorich/console"
	"github.com/depado/gorich/internal/cells"
	"github.com/depado/gorich/markup"
	"github.com/depado/gorich/segment"
	"github.com/depado/gorich/style"
	"github.com/depado/gorich/table/align"
	"github.com/depado/gorich/table/box"
)

// Row holds metadata for a single table row.
type Row struct {
	Style      *style.Style
	EndSection bool
}

// cellRender is an internal type that pairs a renderable with display properties.
type cellRender struct {
	renderable console.Renderable
	style      *style.Style
	vertical   align.VerticalMethod
}

// TableOption configures a Table.
type TableOption func(*Table)

// WithTitle sets the table title.
func WithTitle(title string) TableOption {
	return func(t *Table) { t.title = title }
}

// WithCaption sets the table caption.
func WithCaption(caption string) TableOption {
	return func(t *Table) { t.caption = caption }
}

// WithWidth sets a fixed table width.
func WithWidth(width int) TableOption {
	return func(t *Table) { t.width = width }
}

// WithMinWidth sets the minimum table width.
func WithMinWidth(minWidth int) TableOption {
	return func(t *Table) { t.minWidth = minWidth }
}

// WithBox sets the border box style.
func WithBox(b *box.Box) TableOption {
	return func(t *Table) { t.tableBox = b }
}

// WithExpand enables expanding the table to fill available width.
func WithExpand() TableOption {
	return func(t *Table) { t.expand = true }
}

// WithShowLines enables divider lines between every row.
func WithShowLines() TableOption {
	return func(t *Table) { t.showLines = true }
}

// WithShowHeader sets whether to show the header row.
func WithShowHeader(show bool) TableOption {
	return func(t *Table) { t.showHeader = show }
}

// WithShowFooter sets whether to show the footer row.
func WithShowFooter(show bool) TableOption {
	return func(t *Table) { t.showFooter = show }
}

// WithShowEdge sets whether to draw outer box edges.
func WithShowEdge(show bool) TableOption {
	return func(t *Table) { t.showEdge = show }
}

// WithLeading sets blank lines between rows.
func WithLeading(leading int) TableOption {
	return func(t *Table) { t.leading = leading }
}

// WithStyle sets the default table style.
func WithStyle(s *style.Style) TableOption {
	return func(t *Table) { t.style = s }
}

// WithRowStyles sets alternating row styles.
func WithRowStyles(ss ...*style.Style) TableOption {
	return func(t *Table) { t.rowStyles = ss }
}

// WithHeaderStyle sets the header style.
func WithHeaderStyle(s *style.Style) TableOption {
	return func(t *Table) { t.headerStyle = s }
}

// WithFooterStyle sets the footer style.
func WithFooterStyle(s *style.Style) TableOption {
	return func(t *Table) { t.footerStyle = s }
}

// WithBorderStyle sets the border style.
func WithBorderStyle(s *style.Style) TableOption {
	return func(t *Table) { t.borderStyle = s }
}

// WithPadding sets cell padding.
func WithPad(top, right, bottom, left int) TableOption {
	return func(t *Table) {
		t.padTop = top
		t.padRight = right
		t.padBottom = bottom
		t.padLeft = left
	}
}

// WithCollapsePadding enables padding collapsing between adjacent cells.
func WithCollapsePadding() TableOption {
	return func(t *Table) { t.collapsePadding = true }
}

// WithPadEdge enables padding on edge cells.
func WithPadEdge(pad bool) TableOption {
	return func(t *Table) { t.padEdge = pad }
}

// Table implements a console-renderable table with optional borders, headers,
// footers, and dynamic column width calculation.
type Table struct {
	columns     []*Column
	rows        []Row
	title       string
	caption     string
	tableBox    *box.Box
	showHeader  bool
	showFooter  bool
	showEdge    bool
	showLines   bool
	leading     int
	width       int
	minWidth    int
	expand      bool
	collapsePadding bool
	padEdge     bool

	padTop, padRight, padBottom, padLeft int

	style       *style.Style
	rowStyles   []*style.Style
	headerStyle *style.Style
	footerStyle *style.Style
	borderStyle *style.Style
}

// NewTable creates a new Table with the given column headers.
func NewTable(headers ...string) *Table {
	t := &Table{
		tableBox:    box.HEAVY_HEAD,
		showHeader:  true,
		showFooter:  false,
		showEdge:    true,
		showLines:   false,
		padEdge:     true,
		padRight:    1,
		padLeft:     1,
		expand:      false,
	}

	for i, h := range headers {
		t.columns = append(t.columns, newColumn(h, i))
	}
	return t
}

// NewTableWithOptions creates a Table with headers and options.
func NewTableWithOptions(headers []string, opts ...TableOption) *Table {
	t := NewTable(headers...)
	for _, opt := range opts {
		opt(t)
	}
	return t
}

// AddColumn adds a column to the table with optional column options.
func (t *Table) AddColumn(header string, opts ...ColumnOption) *Column {
	col := newColumn(header, len(t.columns))
	for _, opt := range opts {
		opt(col)
	}
	t.columns = append(t.columns, col)
	return col
}

// AddRow adds a row of values to the table.
func (t *Table) AddRow(values ...interface{}) {
	cols := t.columns
	for i, v := range values {
		if i == len(cols) {
			col := newColumn("", i)
			for range t.rows {
				col.cells = append(col.cells, nil)
			}
			t.columns = append(t.columns, col)
		}
		t.columns[i].cells = append(t.columns[i].cells, t.toRenderable(v))
	}
	for i := len(values); i < len(cols); i++ {
		cols[i].cells = append(cols[i].cells, nil)
	}
	t.rows = append(t.rows, Row{})
}

// AddStyledRow adds a row with an optional row style.
func (t *Table) AddStyledRow(values []interface{}, s *style.Style, endSection bool) {
	for i, v := range values {
		if i == len(t.columns) {
			col := newColumn("", i)
			for range t.rows {
				col.cells = append(col.cells, nil)
			}
			t.columns = append(t.columns, col)
		}
		t.columns[i].cells = append(t.columns[i].cells, t.toRenderable(v))
	}
	for i := len(values); i < len(t.columns); i++ {
		t.columns[i].cells = append(t.columns[i].cells, nil)
	}
	t.rows = append(t.rows, Row{Style: s, EndSection: endSection})
}

// AddSection marks the previous row as the end of a section.
func (t *Table) AddSection() {
	if len(t.rows) > 0 {
		t.rows[len(t.rows)-1].EndSection = true
	}
}

func (t *Table) toRenderable(v interface{}) console.Renderable {
	if v == nil {
		return nil
	}
	switch val := v.(type) {
	case string:
		return &markupRenderable{content: val}
	case console.Renderable:
		return val
	default:
		return &markupRenderable{content: fmt.Sprint(val)}
	}
}

// markupRenderable stores a string and renders it through the markup parser.
type markupRenderable struct {
	content   interface{}
	parsed    *markup.Text
	parsedStr string
}

func (mr *markupRenderable) getParsed() markup.Text {
	s, ok := mr.content.(string)
	if !ok {
		return markup.Text{}
	}
	if mr.parsed != nil && mr.parsedStr == s {
		return *mr.parsed
	}
	t := markup.Parse(s)
	mr.parsed = &t
	mr.parsedStr = s
	return t
}

func (mr *markupRenderable) Render(c *console.Console, opts console.Options) []segment.Segment {
	t := mr.getParsed()
	if t.Plain == "" && len(t.Spans) == 0 {
		_, ok := mr.content.(string)
		if !ok {
			return []segment.Segment{segment.NewText("\n", nil)}
		}
	}
	return t.Render()
}

func (mr *markupRenderable) Measure(c *console.Console, opts console.Options) console.Measurement {
	t := mr.getParsed()
	if t.Plain == "" && len(t.Spans) == 0 {
		_, ok := mr.content.(string)
		if !ok {
			return console.NewMeasurement(0, 0)
		}
	}
	w := cells.Len(t.Plain)
	return console.NewMeasurement(w, w)
}

// Render implements console.Renderable.
func (t *Table) Render(c *console.Console, opts console.Options) []segment.Segment {
	if len(t.columns) == 0 {
		return []segment.Segment{segment.NewText("\n", nil)}
	}

	maxWidth := opts.MaxWidth
	if t.width > 0 {
		maxWidth = t.width
	}

	extraWidth := t.extraWidth()
	calcOpts := opts.WithMaxWidth(maxWidth - extraWidth)
	widths := t.calculateColumnWidths(c, calcOpts)

	tableWidth := sumInts(widths) + extraWidth
	renderOpts := opts.WithMaxWidth(tableWidth)

	var result []segment.Segment

	// Title
	if t.title != "" {
		titleSegs := markup.Render(t.title)
		titleWidth := segment.TotalCellLength(titleSegs)
		titlePad := (tableWidth - titleWidth) / 2
		if titlePad < 0 {
			titlePad = 0
		}
		if titlePad > 0 {
			result = append(result, segment.NewText(spaces(titlePad), nil))
		}
		result = append(result, titleSegs...)
		result = append(result, segment.NewText("\n", nil))
	}

	result = append(result, t._render(c, renderOpts, widths)...)

	// Caption
	if t.caption != "" {
		captionSegs := markup.Render(t.caption)
		captionWidth := segment.TotalCellLength(captionSegs)
		captionPad := (tableWidth - captionWidth) / 2
		if captionPad < 0 {
			captionPad = 0
		}
		if captionPad > 0 {
			result = append(result, segment.NewText(spaces(captionPad), nil))
		}
		result = append(result, captionSegs...)
		result = append(result, segment.NewText("\n", nil))
	}

	return result
}

// Measure implements console.Measurable.
func (t *Table) Measure(c *console.Console, opts console.Options) console.Measurement {
	if len(t.columns) == 0 {
		return console.NewMeasurement(0, 0)
	}

	maxWidth := opts.MaxWidth
	if t.width > 0 {
		maxWidth = t.width
	}
	if maxWidth < 1 {
		return console.NewMeasurement(0, 0)
	}

	extraWidth := t.extraWidth()
	calcOpts := opts.WithMaxWidth(maxWidth - extraWidth)
	widths := t.calculateColumnWidths(c, calcOpts)

	total := sumInts(widths) + extraWidth
	return console.NewMeasurement(total, total)
}

func (t *Table) extraWidth() int {
	// 2 for outer edges + (n-1) for vertical dividers
	if !t.showEdge {
		return len(t.columns) - 1
	}
	return 2 + len(t.columns) - 1
}

func (t *Table) getPadding(firstRow, lastRow bool) (int, int, int, int) {
	top, right, bottom, left := t.padTop, t.padRight, t.padBottom, t.padLeft
	if t.collapsePadding {
		if t.padLeft > t.padRight {
			left = t.padLeft - t.padRight
			right = 0
		} else {
			right = t.padRight - t.padLeft
			left = 0
		}
	}
	if !t.padEdge {
		left = 0
		right = 0
		if firstRow {
			top = 0
		}
		if lastRow {
			bottom = 0
		}
	}
	return top, right, bottom, left
}

func (t *Table) paddingWidth(colIndex int) int {
	_, right, _, left := t.getPadding(colIndex == 0, colIndex == len(t.columns)-1)
	return left + right
}

func spaces(n int) string {
	s := make([]byte, n)
	for i := range s {
		s[i] = ' '
	}
	return string(s)
}
