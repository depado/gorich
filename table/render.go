package table

import (
	"strings"

	"github.com/depado/gorich"
	"github.com/depado/gorich/console"
	"github.com/depado/gorich/segment"
	"github.com/depado/gorich/style"
	"github.com/depado/gorich/table/align"
	"github.com/depado/gorich/table/box"
	"github.com/depado/gorich/table/padding"
)

func (t *Table) _getCells(c *console.Console, colIndex int, col *Column) []cellRender {
	anyPad := t.padTop > 0 || t.padRight > 0 || t.padBottom > 0 || t.padLeft > 0
	firstCol := colIndex == 0
	lastCol := colIndex == len(t.columns)-1

	type rawCell struct {
		style      *style.Style
		renderable console.Renderable
	}
	var rawCells []rawCell

	if t.showHeader {
		var s *style.Style
		if t.headerStyle != nil {
			s = t.headerStyle
		}
		if col.HeaderStyle != nil {
			s = mergeStyles(s, col.HeaderStyle)
		}
		r := t.renderableFor(col.Header)
		rawCells = append(rawCells, rawCell{s, r})
	}

	cellStyle := col.Style
	for _, cell := range col.cells {
		r := cell
		if r == nil {
			r = console.NewText("", nil)
		}
		rawCells = append(rawCells, rawCell{cellStyle, r})
	}

	if t.showFooter {
		var s *style.Style
		if t.footerStyle != nil {
			s = t.footerStyle
		}
		if col.FooterStyle != nil {
			s = mergeStyles(s, col.FooterStyle)
		}
		r := t.renderableFor(col.Footer)
		rawCells = append(rawCells, rawCell{s, r})
	}

	var result []cellRender
	for i, rc := range rawCells {
		firstRow := i == 0
		lastRow := i == len(rawCells)-1

		vert := col.Vertical
		if firstRow && t.showHeader {
			vert = align.Bottom
		} else if lastRow && t.showFooter {
			vert = align.Top
		}

		r := rc.renderable
		if anyPad {
			top, right, bottom, left := t.computePadding(firstCol, lastCol, firstRow, lastRow)
			r = padding.NewPadding(r, []int{top, right, bottom, left}, nil, false)
		}

		result = append(result, cellRender{
			renderable: r,
			style:      rc.style,
			vertical:   vert,
		})
	}
	return result
}

func (t *Table) renderableFor(s string) console.Renderable {
	if s == "" {
		return console.NewText("", nil)
	}
	return gorich.NewMarkupRenderable(s)
}

func (t *Table) computePadding(firstCol, lastCol, firstRow, lastRow bool) (int, int, int, int) {
	top, right, bottom, left := t.padTop, t.padRight, t.padBottom, t.padLeft

	if t.collapsePadding {
		if !firstCol {
			left = maxInt(0, left-right)
		}
		if !lastRow {
			bottom = maxInt(0, top-bottom)
		}
	}

	if !t.padEdge {
		if firstCol {
			left = 0
		}
		if lastCol {
			right = 0
		}
		if firstRow {
			top = 0
		}
		if lastRow {
			bottom = 0
		}
	}

	return top, right, bottom, left
}

func (t *Table) _measureColumn(c *console.Console, opts console.Options, col *Column) console.Measurement {
	optMaxW := opts.MaxWidth
	if optMaxW < 1 {
		return console.NewMeasurement(0, 0)
	}

	padW := t.paddingWidth(col.index)
	if col.Width > 0 {
		return console.NewMeasurement(col.Width+padW, col.Width+padW)
	}

	cells := t._getCells(c, col.index, col)
	var minW, maxW int
	for _, cell := range cells {
		_m, _M := 0, 0
		if m, ok := cell.renderable.(interface {
			Measure(c *console.Console, opts console.Options) console.Measurement
		}); ok {
			meas := m.Measure(c, opts)
			_m = meas.Minimum
			_M = meas.Maximum
		}
		if _m > minW {
			minW = _m
		}
		if _M > maxW {
			maxW = _M
		}
	}
	if minW == 0 {
		minW = 1
	}
	if maxW == 0 {
		maxW = optMaxW
	}

	// Clamp to column min/max
	if col.MinWidth > 0 && minW < col.MinWidth+padW {
		minW = col.MinWidth + padW
	}
	if col.MaxWidth > 0 && maxW > col.MaxWidth+padW {
		maxW = col.MaxWidth + padW
	}

	return console.NewMeasurement(minW, maxW)
}

func (t *Table) calculateColumnWidths(c *console.Console, opts console.Options) []int {
	maxWidth := opts.MaxWidth
	cols := t.columns

	widthRanges := make([]console.Measurement, len(cols))
	for i, col := range cols {
		widthRanges[i] = t._measureColumn(c, opts, col)
	}

	widths := make([]int, len(cols))
	for i, wr := range widthRanges {
		if wr.Maximum > 0 {
			widths[i] = wr.Maximum
		} else {
			widths[i] = 1
		}
	}

	extraWidth := t.extraWidth()

	// Expand with ratios
	if t.expand {
		var ratios []int
		for _, col := range cols {
			if col.flexible() {
				ratios = append(ratios, col.Ratio)
			}
		}
		if len(ratios) > 0 {
			fixedWidths := make([]int, len(cols))
			for i, col := range cols {
				if !col.flexible() {
					fixedWidths[i] = widthRanges[i].Maximum
				}
			}

			var flexMins []int
			for _, col := range cols {
				if col.flexible() {
					w := col.Width
					if w == 0 {
						w = 1
					}
					flexMins = append(flexMins, w+t.paddingWidth(col.index))
				}
			}

			flexibleWidth := maxWidth - sumInts(fixedWidths)
			flexWidths := ratioDistribute(flexibleWidth, ratios, flexMins)

			flexIdx := 0
			for i, col := range cols {
				if col.flexible() {
					widths[i] = fixedWidths[i] + flexWidths[flexIdx]
					flexIdx++
				}
			}
		}
	}

	tableWidth := sumInts(widths)

	// Collapse if too wide
	if tableWidth > maxWidth {
		wrapable := make([]bool, len(cols))
		for i, col := range cols {
			wrapable[i] = col.Width == 0 && !col.NoWrap
		}
		widths = t._collapseWidths(widths, wrapable, maxWidth)
		tableWidth = sumInts(widths)

		if tableWidth > maxWidth {
			excess := tableWidth - maxWidth
			widths = ratioReduce(excess, ones(len(widths)), widths, widths)
		}
	}

	tableWidth = sumInts(widths)

	// Expand if too narrow
	if (tableWidth < maxWidth && t.expand) || (t.minWidth > 0 && tableWidth < t.minWidth-extraWidth) {
		target := maxWidth
		if t.minWidth > 0 {
			mw := t.minWidth - extraWidth
			if mw < target {
				target = mw
			}
		}
		padWidths := ratioDistribute(target-tableWidth, widths, nil)
		for i := range widths {
			widths[i] += padWidths[i]
		}
	}

	return widths
}

func (t *Table) _collapseWidths(widths []int, wrapable []bool, maxWidth int) []int {
	totalWidth := sumInts(widths)
	excess := totalWidth - maxWidth

	anyWrapable := false
	for _, w := range wrapable {
		if w {
			anyWrapable = true
			break
		}
	}
	if !anyWrapable {
		return widths
	}

	for totalWidth > 0 && excess > 0 {
		maxCol := 0
		for i, w := range widths {
			if wrapable[i] && w > maxCol {
				maxCol = w
			}
		}
		secondMax := 0
		for i, w := range widths {
			if wrapable[i] && w != maxCol && w > secondMax {
				secondMax = w
			}
		}
		colDiff := maxCol - secondMax

		ratios := make([]int, len(widths))
		for i := range widths {
			if widths[i] == maxCol && wrapable[i] {
				ratios[i] = 1
			}
		}
		anyRatio := false
		for _, r := range ratios {
			if r > 0 {
				anyRatio = true
				break
			}
		}
		if !anyRatio || colDiff == 0 {
			break
		}

		maxReduce := make([]int, len(widths))
		for i := range maxReduce {
			mr := min(excess, colDiff)
			maxReduce[i] = mr
		}

		widths = ratioReduce(excess, ratios, maxReduce, widths)
		totalWidth = sumInts(widths)
		excess = totalWidth - maxWidth
	}

	return widths
}

func (t *Table) _render(c *console.Console, opts console.Options, widths []int) []segment.Segment {
	b := t.tableBox
	if b == nil {
		b = box.HEAVY_HEAD
	}
	b = b.Substitute(opts)
	if !t.showHeader {
		b = b.GetPlainHeadedBox()
	}

	borderStyle := t.borderStyle

	type boxParts struct {
		left, right, divider segment.Segment
	}

	var boxSegs [3]boxParts
	if b != nil {
		boxSegs[0] = boxParts{
			left:    segment.NewText(string(b.HeadLeft()), borderStyle),
			right:   segment.NewText(string(b.HeadRight()), borderStyle),
			divider: segment.NewText(string(b.HeadVertical()), borderStyle),
		}
		boxSegs[1] = boxParts{
			left:    segment.NewText(string(b.MidLeft()), borderStyle),
			right:   segment.NewText(string(b.MidRight()), borderStyle),
			divider: segment.NewText(string(b.MidVertical()), borderStyle),
		}
		boxSegs[2] = boxParts{
			left:    segment.NewText(string(b.FootLeft()), borderStyle),
			right:   segment.NewText(string(b.FootRight()), borderStyle),
			divider: segment.NewText(string(b.FootVertical()), borderStyle),
		}
	}

	var result []segment.Segment
	nl := segment.NewText("\n", nil)

	// Top border
	if b != nil && t.showEdge {
		topStr := b.GetTop(widths)
		if strings.TrimSpace(topStr) != "" {
			result = append(result, segment.NewText(topStr, borderStyle))
			result = append(result, nl)
		}
	}

	// Build row-oriented cell list
	var rowCells [][]cellRender
	for colIdx, col := range t.columns {
		colCells := t._getCells(c, colIdx, col)
		for rowIdx, cc := range colCells {
			if rowIdx >= len(rowCells) {
				rowCells = append(rowCells, make([]cellRender, len(t.columns)))
			}
			rowCells[rowIdx][colIdx] = cc
		}
	}

	for rowIdx, rowCell := range rowCells {
		first := rowIdx == 0
		last := rowIdx == len(rowCells)-1
		headerRow := first && t.showHeader
		footerRow := last && t.showFooter

		var rowStyle *style.Style
		if !headerRow && !footerRow {
			dataIdx := rowIdx
			if t.showHeader {
				dataIdx--
			}
			if len(t.rowStyles) > 0 && dataIdx >= 0 {
				rowStyle = t.rowStyles[dataIdx%len(t.rowStyles)]
			}
			if dataIdx >= 0 && dataIdx < len(t.rows) && t.rows[dataIdx].Style != nil {
				rowStyle = mergeStyles(rowStyle, t.rows[dataIdx].Style)
			}
		}

		maxHeight := 1
		var cellsLines [][][]segment.Segment

		for colIdx, cell := range rowCell {
			cellOpts := opts.WithMaxWidth(widths[colIdx])
			cellOpts.Justify = t.columns[colIdx].Justify
			cellOpts.NoWrap = t.columns[colIdx].NoWrap
			cellOpts.Overflow = t.columns[colIdx].Overflow

			combinedStyle := cell.style
			if rowStyle != nil {
				combinedStyle = mergeStyles(rowStyle, combinedStyle)
			}

			// Apply style to rendered lines
			lines := c.RenderLines(cell.renderable, cellOpts)
			if combinedStyle != nil {
				for li := range lines {
					lines[li] = segment.ApplyStyle(lines[li], *combinedStyle)
				}
			}
			// Truncate each line to column width
			colW := widths[colIdx]
			for li := range lines {
				lines[li] = segment.AdjustLineLength(lines[li], colW, false)
			}

			if len(lines) > maxHeight {
				maxHeight = len(lines)
			}
			cellsLines = append(cellsLines, lines)
		}

		rowHeight := maxHeight

		// Align and set shape for each cell
		for ci := range cellsLines {
			vert := rowCell[ci].vertical
			if headerRow {
				vert = align.Bottom
			} else if footerRow {
				vert = align.Top
			}

			w := widths[ci]
			bgStyle := rowStyle
			if rowCell[ci].style != nil {
				bgStyle = mergeStyles(bgStyle, rowCell[ci].style)
			}

			aligned := alignVert(cellsLines[ci], vert, w, rowHeight, bgStyle)
			shaped := setShape(aligned, w, maxHeight, bgStyle)
			cellsLines[ci] = shaped
		}

		// Emit row lines
		if b != nil {
			if last && t.showFooter {
				footStr := b.GetRow(widths, "foot")
				if strings.TrimSpace(footStr) != "" {
					result = append(result, segment.NewText(footStr, borderStyle))
					result = append(result, nl)
				}
			}

			var partsIdx int
			if headerRow {
				partsIdx = 0
			} else if footerRow {
				partsIdx = 2
			} else {
				partsIdx = 1
			}

			bp := boxSegs[partsIdx]

			for lineNo := 0; lineNo < maxHeight; lineNo++ {
				if t.showEdge {
					result = append(result, bp.left)
				}
				for ci := range cellsLines {
					if lineNo < len(cellsLines[ci]) {
						result = append(result, cellsLines[ci][lineNo]...)
					} else {
						result = append(result, segment.NewText(spaces(widths[ci]), rowStyle))
					}
					if ci < len(cellsLines)-1 {
						result = append(result, bp.divider)
					}
				}
				if t.showEdge {
					result = append(result, bp.right)
				}
				result = append(result, nl)
			}
		} else {
			for lineNo := 0; lineNo < maxHeight; lineNo++ {
				for ci := range cellsLines {
					if lineNo < len(cellsLines[ci]) {
						result = append(result, cellsLines[ci][lineNo]...)
					} else {
						result = append(result, segment.NewText(spaces(widths[ci]), rowStyle))
					}
				}
				result = append(result, nl)
			}
		}

		// Header divider
		if b != nil && first && t.showHeader {
			headStr := b.GetRow(widths, "head")
			if strings.TrimSpace(headStr) != "" {
				result = append(result, segment.NewText(headStr, borderStyle))
				result = append(result, nl)
			}
		}

		// Row dividers
		endSection := false
		if !headerRow && !footerRow {
			dataIdx := rowIdx
			if t.showHeader {
				dataIdx--
			}
			if dataIdx >= 0 && dataIdx < len(t.rows) && t.rows[dataIdx].EndSection {
				endSection = true
			}
		}

		if b != nil && (t.showLines || t.leading > 0 || endSection) {
			if !last && (!t.showFooter || rowIdx < len(rowCells)-2) && (!t.showHeader || !headerRow) {
				if t.leading > 0 {
					for l := 0; l < t.leading; l++ {
						midStr := b.GetRow(widths, "mid")
						if strings.TrimSpace(midStr) != "" {
							result = append(result, segment.NewText(midStr, borderStyle))
							result = append(result, nl)
						}
					}
				} else {
					rowStr := b.GetRow(widths, "row")
					if strings.TrimSpace(rowStr) != "" {
						result = append(result, segment.NewText(rowStr, borderStyle))
						result = append(result, nl)
					}
				}
			}
		}
	}

	// Bottom border
	if b != nil && t.showEdge {
		bottomStr := b.GetBottom(widths)
		if strings.TrimSpace(bottomStr) != "" {
			result = append(result, segment.NewText(bottomStr, borderStyle))
			result = append(result, nl)
		}
	}

	return result
}

func alignVert(lines [][]segment.Segment, vert align.VerticalMethod, width, height int, style *style.Style) [][]segment.Segment {
	pad := height - len(lines)
	if pad <= 0 {
		return lines
	}

	blankLine := []segment.Segment{segment.NewText(spaces(width), style)}
	switch vert {
	case align.Middle:
		topPad := pad / 2
		bottomPad := pad - topPad
		result := make([][]segment.Segment, 0, height)
		for range topPad {
			result = append(result, blankLine)
		}
		result = append(result, lines...)
		for range bottomPad {
			result = append(result, blankLine)
		}
		return result
	case align.Bottom:
		result := make([][]segment.Segment, 0, height)
		for range pad {
			result = append(result, blankLine)
		}
		result = append(result, lines...)
		return result
	default: // Top
		result := make([][]segment.Segment, 0, height)
		result = append(result, lines...)
		for range pad {
			result = append(result, blankLine)
		}
		return result
	}
}

func setShape(lines [][]segment.Segment, width, height int, style *style.Style) [][]segment.Segment {
	result := make([][]segment.Segment, len(lines))
	for i, line := range lines {
		lw := segment.TotalCellLength(line)
		if lw < width {
			padded := make([]segment.Segment, len(line))
			copy(padded, line)
			padded = append(padded, segment.NewText(spaces(width-lw), style))
			result[i] = padded
		} else {
			result[i] = line
		}
	}
	return result
}

func ones(n int) []int {
	r := make([]int, n)
	for i := range r {
		r[i] = 1
	}
	return r
}

func mergeStyles(base, override *style.Style) *style.Style {
	if override == nil {
		return base
	}
	if base == nil {
		return override
	}
	merged := base.Add(*override)
	return &merged
}
