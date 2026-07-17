package live

import (
	"strings"
	"testing"

	"github.com/depado/gorich/console"
	"github.com/depado/gorich/segment"
)

func newTestConsole() *console.Console {
	return console.New(
		console.WithNoColor(true),
		console.WithForceTerminal(true),
		console.WithWidth(80),
	)
}

func TestBlockDisplayStartAppendFinish(t *testing.T) {
	d := NewBlockDisplay(WithBlockMaxLines(2))
	idx := d.Start("test-task")
	if idx != 0 {
		t.Errorf("expected index 0, got %d", idx)
	}
	d.AppendLine(idx, "line one")
	d.AppendLine(idx, "line two")
	d.AppendLine(idx, "line three") // should be dropped (maxLines=2)

	d.Finish(idx, 0)

	d.mu.Lock()
	blk := d.blocks[idx]
	d.mu.Unlock()

	if blk.Status != BlockSucceeded {
		t.Error("expected BlockSucceeded after Finish(0)")
	}
	if len(blk.Lines) != 2 {
		t.Errorf("expected 2 lines (ring buffer), got %d", len(blk.Lines))
	}
	if blk.Lines[0] != "line two" || blk.Lines[1] != "line three" {
		t.Errorf("expected ['line two', 'line three'], got %v", blk.Lines)
	}
}

func TestBlockDisplayRender(t *testing.T) {
	c := newTestConsole()
	d := NewBlockDisplay(WithBlockMaxLines(2))
	idx := d.Start("task")
	d.AppendLine(idx, "output")
	d.Finish(idx, 0)

	segs := d.Render(c, c.Options())
	text := renderSegsToString(segs)
	if !strings.Contains(text, "✓ task") {
		t.Errorf("expected checkmark header, got: %s", text)
	}
	if !strings.Contains(text, "output") {
		t.Errorf("expected output line, got: %s", text)
	}
}

func renderSegsToString(segs []segment.Segment) string {
	var b strings.Builder
	for _, s := range segs {
		b.WriteString(s.Text)
	}
	return b.String()
}
