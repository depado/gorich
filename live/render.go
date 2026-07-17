// Package live provides auto-refreshing terminal displays.
package live

import (
	"github.com/depado/gorich/console"
	"github.com/depado/gorich/segment"
)

var _ console.Renderable = (*liveRender)(nil)

// VerticalOverflow determines how vertical overflow is handled.
type VerticalOverflow int

const (
	OverflowVisible  VerticalOverflow = iota // Show all content
	OverflowCrop                             // Crop to terminal height
	OverflowEllipsis                         // Show ellipsis for cropped content
)

// liveRender tracks the rendered state and handles cursor repositioning.
type liveRender struct {
	renderable   console.Renderable
	lastHeight   int
	lastWidth    int
	vertOverflow VerticalOverflow
}

// newLiveRender creates a new liveRender.
func newLiveRender(renderable console.Renderable, overflow VerticalOverflow) *liveRender {
	return &liveRender{
		renderable:   renderable,
		vertOverflow: overflow,
	}
}

// SetRenderable updates the renderable to display.
func (lr *liveRender) SetRenderable(r console.Renderable) {
	lr.renderable = r
}

// Reset clears the last rendered shape so the next PositionCursor is a no-op.
// Use after content was committed to scrollback and the live area needs a
// fresh start with no old content to erase.
func (lr *liveRender) Reset() {
	lr.lastHeight = 0
	lr.lastWidth = 0
}

// Shape returns the last rendered shape (width, height).
func (lr *liveRender) Shape() (width, height int) {
	return lr.lastWidth, lr.lastHeight
}

// PositionCursor returns a Control that moves the cursor back to the start
// of the previously rendered content, erasing it in the process.
func (lr *liveRender) PositionCursor() segment.Control {
	if lr.lastHeight == 0 {
		return segment.Control{}
	}

	codes := append(segment.CarriageReturn().Codes, segment.EraseInLine(2).Codes...)
	for i := 1; i < lr.lastHeight; i++ {
		codes = append(codes, segment.CursorUp(1).Codes...)
		codes = append(codes, segment.EraseInLine(2).Codes...)
	}
	return segment.Control{Codes: codes}
}

// RestoreCursor returns a Control that moves the cursor to erase all content
// and returns to the original position. Used for transient mode.
func (lr *liveRender) RestoreCursor() segment.Control {
	if lr.lastHeight == 0 {
		return segment.Control{}
	}

	var codes []segment.ControlCode
	if lr.lastHeight > 1 {
		codes = append(codes, segment.CursorUp(lr.lastHeight-1).Codes...)
	}
	codes = append(codes, segment.CarriageReturn().Codes...)
	codes = append(codes, segment.EraseInDisplay(0).Codes...)
	return segment.Control{Codes: codes}
}

// Render implements console.Renderable.
func (lr *liveRender) Render(c *console.Console, opts console.Options) []segment.Segment {
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
