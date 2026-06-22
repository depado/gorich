package box

import (
	"strings"

	"github.com/depado/gorich/console"
)

// Box defines characters for drawing table borders.
//
// The layout uses 8 lines of 4 characters each:
//
//	┌─┬┐  top
//	│ ││  head
//	├─┼┤  head_row
//	│ ││  mid
//	├─┼┤  row
//	├─┼┤  foot_row
//	│ ││  foot
//	└─┴┘  bottom
type Box struct {
	raw            string
	ascii          bool
	topLeft        rune
	top            rune
	topDivider     rune
	topRight       rune
	headLeft       rune
	headSpace      rune
	headVertical   rune
	headRight      rune
	headRowLeft    rune
	headRowH       rune
	headRowCross   rune
	headRowRight   rune
	midLeft        rune
	midSpace       rune
	midVertical    rune
	midRight       rune
	rowLeft        rune
	rowH           rune
	rowCross       rune
	rowRight       rune
	footRowLeft    rune
	footRowH       rune
	footRowCross   rune
	footRowRight   rune
	footLeft       rune
	footSpace      rune
	footVertical   rune
	footRight      rune
	bottomLeft     rune
	bottom         rune
	bottomDivider  rune
	bottomRight    rune
}

// NewBox creates a Box from an 8-line string of box characters.
func NewBox(raw string, ascii bool) *Box {
	lines := strings.Split(strings.TrimSuffix(raw, "\n"), "\n")
	if len(lines) != 8 {
		panic("box: expected exactly 8 lines")
	}

	b := &Box{raw: raw, ascii: ascii}

	parseLine := func(line string) (rune, rune, rune, rune) {
		runes := []rune(line)
		return runes[0], runes[1], runes[2], runes[3]
	}

	b.topLeft, b.top, b.topDivider, b.topRight = parseLine(lines[0])
	b.headLeft, b.headSpace, b.headVertical, b.headRight = parseLine(lines[1])
	b.headRowLeft, b.headRowH, b.headRowCross, b.headRowRight = parseLine(lines[2])
	b.midLeft, b.midSpace, b.midVertical, b.midRight = parseLine(lines[3])
	b.rowLeft, b.rowH, b.rowCross, b.rowRight = parseLine(lines[4])
	b.footRowLeft, b.footRowH, b.footRowCross, b.footRowRight = parseLine(lines[5])
	b.footLeft, b.footSpace, b.footVertical, b.footRight = parseLine(lines[6])
	b.bottomLeft, b.bottom, b.bottomDivider, b.bottomRight = parseLine(lines[7])

	return b
}

// String returns the raw box definition.
func (b *Box) String() string {
	return b.raw
}

// Substitute returns a different Box if platform constraints prevent this one from rendering.
// Currently only checks for ASCII-only terminals.
func (b *Box) Substitute(opts console.Options) *Box {
	if opts.ASCIIOnly && !b.ascii {
		return ASCII
	}
	return b
}

// GetPlainHeadedBox returns an equivalent box without special header characters.
func (b *Box) GetPlainHeadedBox() *Box {
	if replacement, ok := plainHeadedSubstitutions[b]; ok {
		return replacement
	}
	return b
}

// GetTop builds the top border line string for the given column widths.
func (b *Box) GetTop(widths []int) string {
	var sb strings.Builder
	sb.WriteRune(b.topLeft)
	for i, w := range widths {
		sb.WriteString(strings.Repeat(string(b.top), w))
		if i < len(widths)-1 {
			sb.WriteRune(b.topDivider)
		}
	}
	sb.WriteRune(b.topRight)
	return sb.String()
}

// GetRow builds a horizontal divider line for the given column widths.
// level must be one of "head", "row", "foot", or "mid".
func (b *Box) GetRow(widths []int, level string) string {
	var left, horizontal, cross, right rune
	switch level {
	case "head":
		left, horizontal, cross, right = b.headRowLeft, b.headRowH, b.headRowCross, b.headRowRight
	case "row":
		left, horizontal, cross, right = b.rowLeft, b.rowH, b.rowCross, b.rowRight
	case "mid":
		left, horizontal, cross, right = b.midLeft, ' ', b.midVertical, b.midRight
	case "foot":
		left, horizontal, cross, right = b.footRowLeft, b.footRowH, b.footRowCross, b.footRowRight
	default:
		left, horizontal, cross, right = b.rowLeft, b.rowH, b.rowCross, b.rowRight
	}
	return buildRow(left, horizontal, cross, right, widths)
}

// GetBottom builds the bottom border line string for the given column widths.
func (b *Box) GetBottom(widths []int) string {
	var sb strings.Builder
	sb.WriteRune(b.bottomLeft)
	for i, w := range widths {
		sb.WriteString(strings.Repeat(string(b.bottom), w))
		if i < len(widths)-1 {
			sb.WriteRune(b.bottomDivider)
		}
	}
	sb.WriteRune(b.bottomRight)
	return sb.String()
}

// MidLeft returns the left edge character for a mid (data) row.
func (b *Box) MidLeft() rune { return b.midLeft }

// MidRight returns the right edge character for a mid (data) row.
func (b *Box) MidRight() rune { return b.midRight }

// MidVertical returns the vertical divider character for mid (data) rows.
func (b *Box) MidVertical() rune { return b.midVertical }

// HeadLeft returns the left edge character for the header row.
func (b *Box) HeadLeft() rune { return b.headLeft }

// HeadRight returns the right edge character for the header row.
func (b *Box) HeadRight() rune { return b.headRight }

// HeadVertical returns the vertical divider character for the header row.
func (b *Box) HeadVertical() rune { return b.headVertical }

// FootLeft returns the left edge character for the footer row.
func (b *Box) FootLeft() rune { return b.footLeft }

// FootRight returns the right edge character for the footer row.
func (b *Box) FootRight() rune { return b.footRight }

// FootVertical returns the vertical divider character for the footer row.
func (b *Box) FootVertical() rune { return b.footVertical }

// HeadRowLeft returns the left edge character for a head row divider.
func (b *Box) HeadRowLeft() rune { return b.headRowLeft }

// HeadRowH returns the horizontal character for a head row divider.
func (b *Box) HeadRowH() rune { return b.headRowH }

// HeadRowCross returns the cross character for a head row divider.
func (b *Box) HeadRowCross() rune { return b.headRowCross }

// HeadRowRight returns the right edge character for a head row divider.
func (b *Box) HeadRowRight() rune { return b.headRowRight }

// NoEdge returns a Box with outer edges replaced by spaces.
func (b *Box) NoEdge() *Box {
	c := *b
	c.topLeft, c.topRight = ' ', ' '
	c.headLeft, c.headRight = ' ', ' '
	c.midLeft, c.midRight = ' ', ' '
	c.footLeft, c.footRight = ' ', ' '
	c.bottomLeft, c.bottomRight = ' ', ' '
	return &c
}

func buildRow(left, horizontal, divider, right rune, widths []int) string {
	var sb strings.Builder
	sb.WriteRune(left)
	for i, w := range widths {
		sb.WriteString(strings.Repeat(string(horizontal), w))
		if i < len(widths)-1 {
			sb.WriteRune(divider)
		}
	}
	sb.WriteRune(right)
	return sb.String()
}

// Predefined box styles.

var (
	ASCII = NewBox(
		"+--+\n"+
			"| ||\n"+
			"|-+|\n"+
			"| ||\n"+
			"|-+|\n"+
			"|-+|\n"+
			"| ||\n"+
			"+--+\n",
		true,
	)

	ASCII2 = NewBox(
		"+-++\n"+
			"| ||\n"+
			"+-++\n"+
			"| ||\n"+
			"+-++\n"+
			"+-++\n"+
			"| ||\n"+
			"+-++\n",
		true,
	)

	ASCII_DOUBLE_HEAD = NewBox(
		"+-++\n"+
			"| ||\n"+
			"+=++\n"+
			"| ||\n"+
			"+-++\n"+
			"+-++\n"+
			"| ||\n"+
			"+-++\n",
		true,
	)

	SQUARE = NewBox(
		"┌─┬┐\n"+
			"│ ││\n"+
			"├─┼┤\n"+
			"│ ││\n"+
			"├─┼┤\n"+
			"├─┼┤\n"+
			"│ ││\n"+
			"└─┴┘\n",
		false,
	)

	SQUARE_DOUBLE_HEAD = NewBox(
		"┌─┬┐\n"+
			"│ ││\n"+
			"╞═╪╡\n"+
			"│ ││\n"+
			"├─┼┤\n"+
			"├─┼┤\n"+
			"│ ││\n"+
			"└─┴┘\n",
		false,
	)

	MINIMAL = NewBox(
		"  ╷ \n"+
			"  │ \n"+
			"╶─┼╴\n"+
			"  │ \n"+
			"╶─┼╴\n"+
			"╶─┼╴\n"+
			"  │ \n"+
			"  ╵ \n",
		false,
	)

	MINIMAL_HEAVY_HEAD = NewBox(
		"  ╷ \n"+
			"  │ \n"+
			"╺━┿╸\n"+
			"  │ \n"+
			"╶─┼╴\n"+
			"╶─┼╴\n"+
			"  │ \n"+
			"  ╵ \n",
		false,
	)

	MINIMAL_DOUBLE_HEAD = NewBox(
		"  ╷ \n"+
			"  │ \n"+
			" ═╪ \n"+
			"  │ \n"+
			" ─┼ \n"+
			" ─┼ \n"+
			"  │ \n"+
			"  ╵ \n",
		false,
	)

	SIMPLE = NewBox(
		"    \n"+
			"    \n"+
			" ── \n"+
			"    \n"+
			"    \n"+
			" ── \n"+
			"    \n"+
			"    \n",
		false,
	)

	SIMPLE_HEAD = NewBox(
		"    \n"+
			"    \n"+
			" ── \n"+
			"    \n"+
			"    \n"+
			"    \n"+
			"    \n"+
			"    \n",
		false,
	)

	SIMPLE_HEAVY = NewBox(
		"    \n"+
			"    \n"+
			" ━━ \n"+
			"    \n"+
			"    \n"+
			" ━━ \n"+
			"    \n"+
			"    \n",
		false,
	)

	HORIZONTALS = NewBox(
		" ── \n"+
			"    \n"+
			" ── \n"+
			"    \n"+
			" ── \n"+
			" ── \n"+
			"    \n"+
			" ── \n",
		false,
	)

	ROUNDED = NewBox(
		"╭─┬╮\n"+
			"│ ││\n"+
			"├─┼┤\n"+
			"│ ││\n"+
			"├─┼┤\n"+
			"├─┼┤\n"+
			"│ ││\n"+
			"╰─┴╯\n",
		false,
	)

	HEAVY = NewBox(
		"┏━┳┓\n"+
			"┃ ┃┃\n"+
			"┣━╋┫\n"+
			"┃ ┃┃\n"+
			"┣━╋┫\n"+
			"┣━╋┫\n"+
			"┃ ┃┃\n"+
			"┗━┻┛\n",
		false,
	)

	HEAVY_EDGE = NewBox(
		"┏━┯┓\n"+
			"┃ │┃\n"+
			"┠─┼┨\n"+
			"┃ │┃\n"+
			"┠─┼┨\n"+
			"┠─┼┨\n"+
			"┃ │┃\n"+
			"┗━┷┛\n",
		false,
	)

	HEAVY_HEAD = NewBox(
		"┏━┳┓\n"+
			"┃ ┃┃\n"+
			"┡━╇┩\n"+
			"│ ││\n"+
			"├─┼┤\n"+
			"├─┼┤\n"+
			"│ ││\n"+
			"└─┴┘\n",
		false,
	)

	DOUBLE = NewBox(
		"╔═╦╗\n"+
			"║ ║║\n"+
			"╠═╬╣\n"+
			"║ ║║\n"+
			"╠═╬╣\n"+
			"╠═╬╣\n"+
			"║ ║║\n"+
			"╚═╩╝\n",
		false,
	)

	DOUBLE_EDGE = NewBox(
		"╔═╤╗\n"+
			"║ │║\n"+
			"╟─┼╢\n"+
			"║ │║\n"+
			"╟─┼╢\n"+
			"╟─┼╢\n"+
			"║ │║\n"+
			"╚═╧╝\n",
		false,
	)

	MARKDOWN = NewBox(
		"    \n"+
			"| ||\n"+
			"|-||\n"+
			"| ||\n"+
			"|-||\n"+
			"|-||\n"+
			"| ||\n"+
			"    \n",
		true,
	)
)

var plainHeadedSubstitutions = map[*Box]*Box{
	HEAVY_HEAD:           SQUARE,
	SQUARE_DOUBLE_HEAD:   SQUARE,
	MINIMAL_DOUBLE_HEAD:  MINIMAL,
	MINIMAL_HEAVY_HEAD:   MINIMAL,
	ASCII_DOUBLE_HEAD:    ASCII2,
}
