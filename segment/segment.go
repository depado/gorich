package segment

import (
	"strings"

	"github.com/depado/gorich/internal/cells"
	"github.com/depado/gorich/style"
)

// Segment is the atomic unit of terminal rendering.
// A segment is either styled text or a control sequence.
type Segment struct {
	Text    string
	Style   *style.Style
	Control []ControlCode // Non-nil means this is a control segment
}

// NewText creates a text segment with optional style.
func NewText(text string, s *style.Style) Segment {
	return Segment{Text: text, Style: s}
}

// NewControl creates a control segment.
func NewControlSegment(codes ...ControlCode) Segment {
	text := ""
	for _, c := range codes {
		text += c.Render()
	}
	return Segment{Text: text, Control: codes}
}

// IsControl returns true if this is a control segment.
func (s Segment) IsControl() bool {
	return s.Control != nil
}

// CellLength returns the number of terminal cells this segment occupies.
// Control segments have zero cell length.
func (s Segment) CellLength() int {
	if s.IsControl() {
		return 0
	}
	return cells.Len(s.Text)
}

// Render produces the ANSI-encoded string for this segment.
func (s Segment) Render(colorSystem style.ColorSystem) string {
	if s.IsControl() {
		return s.Text // Control codes are pre-rendered
	}
	if s.Style == nil {
		return s.Text
	}
	return s.Style.Render(s.Text, colorSystem)
}

// SplitLines splits segments by newlines, returning lines of segments.
func SplitLines(segments []Segment) [][]Segment {
	if len(segments) == 0 {
		return nil
	}

	var lines [][]Segment
	var currentLine []Segment

	for _, seg := range segments {
		if seg.IsControl() {
			currentLine = append(currentLine, seg)
			continue
		}

		parts := strings.Split(seg.Text, "\n")
		for i, part := range parts {
			if i > 0 {
				lines = append(lines, currentLine)
				currentLine = nil
			}
			if part != "" || i == len(parts)-1 {
				currentLine = append(currentLine, Segment{
					Text:  part,
					Style: seg.Style,
				})
			}
		}
	}

	if len(currentLine) > 0 {
		lines = append(lines, currentLine)
	}

	return lines
}

// AdjustLineLength adjusts a line of segments to the given cell width.
// If pad is true, shorter lines are padded with spaces.
// Longer lines are truncated.
func AdjustLineLength(line []Segment, width int, pad bool) []Segment {
	if width <= 0 {
		return nil
	}

	currentLen := 0
	for _, seg := range line {
		currentLen += seg.CellLength()
	}

	if currentLen == width {
		return line
	}

	if currentLen < width {
		if pad {
			padding := width - currentLen
			line = append(line, Segment{Text: strings.Repeat(" ", padding)})
		}
		return line
	}

	// Need to truncate
	var result []Segment
	remaining := width

	for _, seg := range line {
		if seg.IsControl() {
			result = append(result, seg)
			continue
		}

		segLen := seg.CellLength()
		if segLen <= remaining {
			result = append(result, seg)
			remaining -= segLen
		} else if remaining > 0 {
			// Truncate this segment
			truncated := truncateToWidth(seg.Text, remaining)
			result = append(result, Segment{Text: truncated, Style: seg.Style})
			remaining = 0
		}
		if remaining == 0 {
			break
		}
	}

	return result
}

// truncateToWidth truncates a string to fit within the given cell width.
func truncateToWidth(s string, width int) string {
	if width <= 0 {
		return ""
	}

	currentWidth := 0
	for i, r := range s {
		runeWidth := cells.RuneWidth(r)
		if currentWidth+runeWidth > width {
			return s[:i]
		}
		currentWidth += runeWidth
	}
	return s
}

// TotalCellLength returns the total cell length of a slice of segments.
func TotalCellLength(segments []Segment) int {
	total := 0
	for _, seg := range segments {
		total += seg.CellLength()
	}
	return total
}

// Simplify merges adjacent segments with the same style.
func Simplify(segments []Segment) []Segment {
	if len(segments) <= 1 {
		return segments
	}

	var result []Segment
	for _, seg := range segments {
		if seg.IsControl() || len(result) == 0 {
			result = append(result, seg)
			continue
		}

		last := &result[len(result)-1]
		if last.IsControl() {
			result = append(result, seg)
			continue
		}

		// Check if styles are the same
		if stylesEqual(last.Style, seg.Style) {
			last.Text += seg.Text
		} else {
			result = append(result, seg)
		}
	}

	return result
}

func stylesEqual(a, b *style.Style) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	// Simple equality check - could be more thorough
	return *a == *b
}

// ApplyStyle applies a style to all segments.
func ApplyStyle(segments []Segment, s style.Style) []Segment {
	result := make([]Segment, len(segments))
	for i, seg := range segments {
		if seg.IsControl() {
			result[i] = seg
		} else {
			newStyle := s
			if seg.Style != nil {
				newStyle = s.Add(*seg.Style)
			}
			result[i] = Segment{Text: seg.Text, Style: &newStyle}
		}
	}
	return result
}

// Divide divides segments at the given cell positions.
// Returns len(cuts)+1 slices of segments.
func Divide(segments []Segment, cuts []int) [][]Segment {
	if len(cuts) == 0 {
		return [][]Segment{segments}
	}

	result := make([][]Segment, len(cuts)+1)
	currentPos := 0
	cutIdx := 0
	segIdx := 0

	for cutIdx < len(cuts) {
		cut := cuts[cutIdx]
		var currentSlice []Segment

		for segIdx < len(segments) && currentPos < cut {
			seg := segments[segIdx]
			if seg.IsControl() {
				currentSlice = append(currentSlice, seg)
				segIdx++
				continue
			}

			segLen := seg.CellLength()
			segEnd := currentPos + segLen

			if segEnd <= cut {
				currentSlice = append(currentSlice, seg)
				currentPos = segEnd
				segIdx++
			} else {
				// Need to split this segment
				splitWidth := cut - currentPos
				left := truncateToWidth(seg.Text, splitWidth)
				right := seg.Text[len(left):]

				currentSlice = append(currentSlice, Segment{Text: left, Style: seg.Style})
				segments[segIdx] = Segment{Text: right, Style: seg.Style}
				currentPos = cut
			}
		}

		result[cutIdx] = currentSlice
		cutIdx++
	}

	// Remaining segments go in the last slice
	result[len(cuts)] = segments[segIdx:]

	return result
}
