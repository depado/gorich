package padding

import (
	"strings"

	"github.com/depado/gorich/console"
	"github.com/depado/gorich/segment"
	"github.com/depado/gorich/style"
)

// Padding is a renderable that wraps content with CSS-style padding.
type Padding struct {
	renderable console.Renderable
	top        int
	right      int
	bottom     int
	left       int
	style      *style.Style
	expand     bool
}

// NewPadding creates a Padding renderable.
func NewPadding(renderable console.Renderable, pad []int, style *style.Style, expand bool) *Padding {
	top, right, bottom, left := unpack(pad)
	return &Padding{
		renderable: renderable,
		top:        top,
		right:      right,
		bottom:     bottom,
		left:       left,
		style:      style,
		expand:     expand,
	}
}

// NewIndent creates a Padding that indents content by the given level.
func NewIndent(renderable console.Renderable, level int) *Padding {
	return NewPadding(renderable, []int{0, 0, 0, level}, nil, false)
}

func unpack(pad []int) (int, int, int, int) {
	switch len(pad) {
	case 1:
		return pad[0], pad[0], pad[0], pad[0]
	case 2:
		return pad[0], pad[1], pad[0], pad[1]
	case 4:
		return pad[0], pad[1], pad[2], pad[3]
	default:
		return 0, 0, 0, 0
	}
}

// Render implements console.Renderable.
func (p *Padding) Render(c *console.Console, opts console.Options) []segment.Segment {
	lines := p.contentLines(c, opts)

	var result []segment.Segment

	// Top padding: blank lines at full width
	if p.top > 0 {
		fullWidth := p.fullWidth(lines, opts.MaxWidth)
		if fullWidth > 0 {
			var blankLine []segment.Segment
			blankLine = append(blankLine, segment.NewText(strings.Repeat(" ", fullWidth), p.style))
			for i := 0; i < p.top; i++ {
				if i > 0 || len(result) > 0 {
					result = append(result, segment.NewText("\n", nil))
				}
				result = append(result, blankLine...)
			}
		}
	}

	// Content lines with left/right padding
	for i, line := range lines {
		if i > 0 || (p.top > 0 && len(result) > 0) {
			result = append(result, segment.NewText("\n", nil))
		}
		if p.left > 0 {
			result = append(result, segment.NewText(strings.Repeat(" ", p.left), p.style))
		}
		result = append(result, line...)
		if p.right > 0 {
			result = append(result, segment.NewText(strings.Repeat(" ", p.right), p.style))
		}
	}

	// Bottom padding: blank lines at full width
	if p.bottom > 0 {
		fullWidth := p.fullWidth(lines, opts.MaxWidth)
		if fullWidth > 0 {
			for i := 0; i < p.bottom; i++ {
				result = append(result, segment.NewText("\n", nil))
				result = append(result, segment.NewText(strings.Repeat(" ", fullWidth), p.style))
			}
		}
	}

	return result
}

func (p *Padding) contentLines(c *console.Console, opts console.Options) [][]segment.Segment {
	var width int
	if p.expand {
		width = opts.MaxWidth
	} else {
		// Measure the renderable, or render it and measure the result
		meas := console.NewMeasurement(0, 0)
		if m, ok := p.renderable.(interface {
			Measure(c *console.Console, opts console.Options) console.Measurement
		}); ok {
			meas = m.Measure(c, opts)
		} else {
			// Render to find actual width
			segs := p.renderable.Render(c, opts)
			w := segment.TotalCellLength(segs)
			meas = console.NewMeasurement(w, w)
		}
		width = min(meas.Maximum+p.left+p.right, opts.MaxWidth)
		if width < 1 {
			return nil
		}
	}

	innerWidth := width - p.left - p.right
	if innerWidth < 1 {
		return nil
	}

	innerOpts := opts.WithMaxWidth(innerWidth)
	segs := p.renderable.Render(c, innerOpts)
	return segment.SplitLines(segs)
}

func (p *Padding) fullWidth(lines [][]segment.Segment, maxWidth int) int {
	if p.expand {
		return maxWidth
	}
	max := 0
	for _, line := range lines {
		w := segment.TotalCellLength(line) + p.left + p.right
		if w > max {
			max = w
		}
	}
	return max
}

// Measure implements console.Measurable.
func (p *Padding) Measure(c *console.Console, opts console.Options) console.Measurement {
	maxWidth := opts.MaxWidth
	extra := p.left + p.right
	if maxWidth-extra < 1 {
		return console.NewMeasurement(maxWidth, maxWidth)
	}

	var measureMin, measureMax int
	if m, ok := p.renderable.(interface {
		Measure(c *console.Console, opts console.Options) console.Measurement
	}); ok {
		meas := m.Measure(c, opts)
		measureMin = meas.Minimum
		measureMax = meas.Maximum
	}

	result := console.NewMeasurement(measureMin+extra, measureMax+extra)
	if result.Maximum > maxWidth {
		result.Maximum = maxWidth
	}
	return result
}
