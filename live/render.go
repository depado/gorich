// Package live provides auto-refreshing terminal displays.
package live

import (
	"github.com/depado/gorich/console"
	"github.com/depado/gorich/segment"
)

// VerticalOverflow determines how vertical overflow is handled.
type VerticalOverflow int

const (
	OverflowVisible VerticalOverflow = iota // Show all content
	OverflowCrop                            // Crop to terminal height
	OverflowEllipsis                        // Show ellipsis for cropped content
)

// LiveRender tracks the rendered state and handles cursor repositioning.
type LiveRender struct {
	renderable   console.Renderable
	lastHeight   int
	lastWidth    int
	vertOverflow VerticalOverflow
}

// NewLiveRender creates a new LiveRender.
func NewLiveRender(renderable console.Renderable, overflow VerticalOverflow) *LiveRender {
	return &LiveRender{
		renderable:   renderable,
		vertOverflow: overflow,
	}
}

// SetRenderable updates the renderable to display.
func (lr *LiveRender) SetRenderable(r console.Renderable) {
	lr.renderable = r
}

// Shape returns the last rendered shape (width, height).
func (lr *LiveRender) Shape() (width, height int) {
	return lr.lastWidth, lr.lastHeight
}

// PositionCursor returns a Control that moves the cursor back to the start
// of the previously rendered content, erasing it in the process.
func (lr *LiveRender) PositionCursor() segment.Control {
	if lr.lastHeight == 0 {
		return segment.Control{}
	}

	// Build control codes:
	// 1. Carriage return (go to column 0)
	// 2. For each line: erase line + cursor up
	var codes []segment.ControlCode

	// Carriage return
	codes = append(codes, segment.ControlCode{Type: segment.ControlCarriageReturn})

	// Erase current line
	codes = append(codes, segment.ControlCode{Type: segment.ControlEraseInLine, Params: []int{2}})

	// Move up and erase each previous line
	for i := 1; i < lr.lastHeight; i++ {
		codes = append(codes, segment.ControlCode{Type: segment.ControlCursorUp, Params: []int{1}})
		codes = append(codes, segment.ControlCode{Type: segment.ControlEraseInLine, Params: []int{2}})
	}

	return segment.Control{Codes: codes}
}

// RestoreCursor returns a Control that moves the cursor to erase all content
// and returns to the original position. Used for transient mode.
func (lr *LiveRender) RestoreCursor() segment.Control {
	if lr.lastHeight == 0 {
		return segment.Control{}
	}

	// Move up to the start, then erase from cursor to end of display
	var codes []segment.ControlCode

	// Move up to the start of the live content
	if lr.lastHeight > 1 {
		codes = append(codes, segment.ControlCode{Type: segment.ControlCursorUp, Params: []int{lr.lastHeight - 1}})
	}

	// Carriage return
	codes = append(codes, segment.ControlCode{Type: segment.ControlCarriageReturn})

	// Erase from cursor to end of display
	codes = append(codes, segment.ControlCode{Type: segment.ControlEraseInDisplay, Params: []int{0}})

	return segment.Control{Codes: codes}
}

// Render implements console.Renderable.
func (lr *LiveRender) Render(c *console.Console, opts console.Options) []segment.Segment {
	if lr.renderable == nil {
		lr.lastHeight = 0
		lr.lastWidth = 0
		return nil
	}

	// Render the content
	segments := lr.renderable.Render(c, opts)

	// Split into lines and calculate shape
	lines := segment.SplitLines(segments)
	height := len(lines)

	// Apply vertical overflow handling
	maxHeight := opts.MaxHeight
	if maxHeight > 0 && lr.vertOverflow != OverflowVisible && height > maxHeight {
		switch lr.vertOverflow {
		case OverflowCrop:
			lines = lines[:maxHeight]
			height = maxHeight
		case OverflowEllipsis:
			lines = lines[:maxHeight-1]
			lines = append(lines, []segment.Segment{{Text: "..."}})
			height = maxHeight
		}
	}

	// Calculate width
	width := 0
	for _, line := range lines {
		lineWidth := segment.TotalCellLength(line)
		if lineWidth > width {
			width = lineWidth
		}
	}

	lr.lastHeight = height
	lr.lastWidth = width

	// Reconstruct segments from lines (add newlines between)
	var result []segment.Segment
	for i, line := range lines {
		result = append(result, line...)
		if i < len(lines)-1 {
			result = append(result, segment.Segment{Text: "\n"})
		}
	}

	return result
}
