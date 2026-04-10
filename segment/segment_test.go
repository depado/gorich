package segment

import (
	"testing"

	"github.com/depado/gorich/style"
)

func TestSegmentCellLength(t *testing.T) {
	tests := []struct {
		name string
		seg  Segment
		want int
	}{
		{"empty", Segment{Text: ""}, 0},
		{"ascii", Segment{Text: "hello"}, 5},
		{"control", Segment{Text: "\x1b[1m", Control: []ControlCode{{Type: ControlCursorUp}}}, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.seg.CellLength()
			if got != tt.want {
				t.Errorf("CellLength() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestSegmentIsControl(t *testing.T) {
	text := Segment{Text: "hello"}
	if text.IsControl() {
		t.Error("Text segment should not be control")
	}

	ctrl := Segment{Text: "\x1b[1A", Control: []ControlCode{{Type: ControlCursorUp}}}
	if !ctrl.IsControl() {
		t.Error("Control segment should be control")
	}
}

func TestSplitLines(t *testing.T) {
	segments := []Segment{
		{Text: "line1\nline2\nline3"},
	}

	lines := SplitLines(segments)
	if len(lines) != 3 {
		t.Errorf("SplitLines produced %d lines, want 3", len(lines))
	}

	if lines[0][0].Text != "line1" {
		t.Errorf("Line 0 = %q, want %q", lines[0][0].Text, "line1")
	}
	if lines[1][0].Text != "line2" {
		t.Errorf("Line 1 = %q, want %q", lines[1][0].Text, "line2")
	}
	if lines[2][0].Text != "line3" {
		t.Errorf("Line 2 = %q, want %q", lines[2][0].Text, "line3")
	}
}

func TestAdjustLineLength(t *testing.T) {
	line := []Segment{{Text: "hello"}}

	// Pad to 10
	padded := AdjustLineLength(line, 10, true)
	total := TotalCellLength(padded)
	if total != 10 {
		t.Errorf("Padded length = %d, want 10", total)
	}

	// Truncate to 3
	truncated := AdjustLineLength(line, 3, false)
	total = TotalCellLength(truncated)
	if total != 3 {
		t.Errorf("Truncated length = %d, want 3", total)
	}
}

func TestSimplify(t *testing.T) {
	s := style.New().WithBold(true)
	segments := []Segment{
		{Text: "hello", Style: &s},
		{Text: " ", Style: &s},
		{Text: "world", Style: &s},
	}

	simplified := Simplify(segments)
	if len(simplified) != 1 {
		t.Errorf("Simplified to %d segments, want 1", len(simplified))
	}
	if simplified[0].Text != "hello world" {
		t.Errorf("Simplified text = %q, want %q", simplified[0].Text, "hello world")
	}
}

func TestTotalCellLength(t *testing.T) {
	segments := []Segment{
		{Text: "hello"},
		{Text: " "},
		{Text: "world"},
	}

	total := TotalCellLength(segments)
	if total != 11 {
		t.Errorf("TotalCellLength = %d, want 11", total)
	}
}

func TestControlRender(t *testing.T) {
	tests := []struct {
		name string
		ctrl Control
		want string
	}{
		{"cursor up 1", CursorUp(1), "\x1b[1A"},
		{"cursor up 5", CursorUp(5), "\x1b[5A"},
		{"cursor down", CursorDown(2), "\x1b[2B"},
		{"carriage return", CarriageReturn(), "\r"},
		{"erase line", EraseInLine(2), "\x1b[2K"},
		{"hide cursor", HideCursor(), "\x1b[?25l"},
		{"show cursor", ShowCursor(), "\x1b[?25h"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.ctrl.Render()
			if got != tt.want {
				t.Errorf("Render() = %q, want %q", got, tt.want)
			}
		})
	}
}
