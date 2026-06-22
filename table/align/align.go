package align

import (
	"strings"

	"github.com/depado/gorich/console"
	"github.com/depado/gorich/segment"
	"github.com/depado/gorich/style"
)

// Method represents horizontal alignment.
type Method string

const (
	Left   Method = "left"
	Center Method = "center"
	Right  Method = "right"
)

// VerticalMethod represents vertical alignment.
type VerticalMethod string

const (
	Top    VerticalMethod = "top"
	Middle VerticalMethod = "middle"
	Bottom VerticalMethod = "bottom"
)

// Align aligns a renderable horizontally and vertically within a given space.
type Align struct {
	renderable console.Renderable
	align      Method
	vertical   VerticalMethod
	style      *style.Style
	pad        bool
	width      int // 0 = auto from content
	height     int // 0 = auto from content
}

// New creates an Align renderable.
func New(renderable console.Renderable, align Method, vertical VerticalMethod, style *style.Style, pad bool, width, height int) *Align {
	return &Align{
		renderable: renderable,
		align:      align,
		vertical:   vertical,
		style:      style,
		pad:        pad,
		width:      width,
		height:     height,
	}
}

// Render implements console.Renderable.
func (a *Align) Render(c *console.Console, opts console.Options) []segment.Segment {
	content := a.renderable.Render(c, opts)
	lines := segment.SplitLines(content)

	width := a.width
	if width == 0 {
		for _, line := range lines {
			lw := segment.TotalCellLength(line)
			if lw > width {
				width = lw
			}
		}
	}
	if width == 0 {
		return nil
	}

	height := a.height
	if height == 0 {
		height = len(lines)
	}

	// Horizontal alignment: pad each line
	aligned := make([][]segment.Segment, len(lines))
	for i, line := range lines {
		lineWidth := segment.TotalCellLength(line)
		padding := width - lineWidth
		var padded []segment.Segment

		switch a.align {
		case Right:
			if padding > 0 && a.pad {
				padded = append(padded, segment.NewText(strings.Repeat(" ", padding), a.style))
			}
			padded = append(padded, line...)
		case Center:
			leftPad := padding / 2
			rightPad := padding - leftPad
			if leftPad > 0 && a.pad {
				padded = append(padded, segment.NewText(strings.Repeat(" ", leftPad), a.style))
			}
			padded = append(padded, line...)
			if rightPad > 0 && a.pad {
				padded = append(padded, segment.NewText(strings.Repeat(" ", rightPad), a.style))
			}
		default: // Left
			padded = append(padded, line...)
			if padding > 0 && a.pad {
				padded = append(padded, segment.NewText(strings.Repeat(" ", padding), a.style))
			}
		}

		aligned[i] = padded
	}

	// Vertical alignment
	topPad, bottomPad := 0, 0
	if height > len(aligned) {
		switch a.vertical {
		case Middle:
			topPad = (height - len(aligned)) / 2
			bottomPad = height - len(aligned) - topPad
		case Bottom:
			topPad = height - len(aligned)
		default: // Top
			bottomPad = height - len(aligned)
		}
	}

	var result []segment.Segment
	blankLine := segment.NewText(strings.Repeat(" ", width), a.style)
	newline := segment.NewText("\n", nil)

	for i := 0; i < topPad; i++ {
		result = append(result, blankLine)
		result = append(result, newline)
	}

	for i, line := range aligned {
		result = append(result, line...)
		if i < len(aligned)-1 || bottomPad > 0 {
			result = append(result, newline)
		}
	}

	for i := 0; i < bottomPad; i++ {
		result = append(result, blankLine)
		if i < bottomPad-1 {
			result = append(result, newline)
		}
	}

	return result
}
