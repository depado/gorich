package gorich

import (
	"testing"

	"github.com/depado/gorich/console"
	"github.com/depado/gorich/markup"
)

func TestMarkupRenderable(t *testing.T) {
	c := console.New(console.WithNoColor(true), console.WithForceTerminal(true))
	mr := NewMarkupRenderable("[bold]Hello[/] World")

	segs := mr.Render(c, c.Options())
	if len(segs) == 0 {
		t.Error("expected non-empty segments")
	}
	totalLen := 0
	for _, seg := range segs {
		totalLen += len(seg.Text)
	}
	if totalLen < 11 {
		t.Errorf("expected at least 11 characters, got %d", totalLen)
	}

	meas := mr.Measure(c, c.Options())
	if meas.Minimum != meas.Maximum {
		t.Errorf("expected same min/max for plain text, got %d/%d", meas.Minimum, meas.Maximum)
	}
	if meas.Maximum < 11 {
		t.Errorf("expected max >= 11, got %d", meas.Maximum)
	}

	if mr.Content() != "[bold]Hello[/] World" {
		t.Error("Content() should return original markup string")
	}
}

func TestMarkupText(t *testing.T) {
	c := console.New(console.WithNoColor(true), console.WithForceTerminal(true))

	parsed := markup.Parse("[bold]Hello[/] World")
	mt := NewMarkupText(parsed)

	segs := mt.Render(c, c.Options())
	if len(segs) == 0 {
		t.Error("expected non-empty segments")
	}
	totalLen := 0
	for _, seg := range segs {
		totalLen += len(seg.Text)
	}
	if totalLen < 11 {
		t.Errorf("expected at least 11 characters, got %d", totalLen)
	}

	meas := mt.Measure(c, c.Options())
	if meas.Minimum != meas.Maximum {
		t.Errorf("expected same min/max, got %d/%d", meas.Minimum, meas.Maximum)
	}
	if meas.Maximum < 11 {
		t.Errorf("expected max >= 11, got %d", meas.Maximum)
	}

	var _ console.Renderable = mt
	var _ console.Measurable = mt
}
