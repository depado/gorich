package gorich

import (
	"sync"

	"github.com/depado/gorich/console"
	"github.com/depado/gorich/internal/cells"
	"github.com/depado/gorich/markup"
	"github.com/depado/gorich/segment"
)

// MarkupRenderable wraps a string as a console.Renderable that parses
// Rich-style markup on render. It implements both console.Renderable
// and console.Measurable.
type MarkupRenderable struct {
	mu        sync.Mutex
	content   string
	parsed    *markup.Text
	parsedStr string
}

// NewMarkupRenderable creates a MarkupRenderable from a markup string.
func NewMarkupRenderable(content string) *MarkupRenderable {
	return &MarkupRenderable{content: content}
}

func (mr *MarkupRenderable) getParsed() markup.Text {
	mr.mu.Lock()
	defer mr.mu.Unlock()
	if mr.parsed != nil && mr.parsedStr == mr.content {
		return *mr.parsed
	}
	t := markup.Parse(mr.content)
	mr.parsed = &t
	mr.parsedStr = mr.content
	return t
}

// Render implements console.Renderable.
func (mr *MarkupRenderable) Render(c *console.Console, opts console.Options) []segment.Segment {
	return mr.getParsed().Render()
}

// Measure implements console.Measurable.
func (mr *MarkupRenderable) Measure(c *console.Console, opts console.Options) console.Measurement {
	w := cells.Len(mr.getParsed().Plain)
	return console.NewMeasurement(w, w)
}

// Content returns the raw markup string.
func (mr *MarkupRenderable) Content() string {
	return mr.content
}

// MarkupText wraps a pre-parsed markup.Text as a console.Renderable.
// Use this to avoid re-parsing when you already have a parsed result.
// For raw markup strings, use NewMarkupRenderable instead.
type MarkupText struct {
	Text markup.Text
}

// NewMarkupText creates a MarkupText from a pre-parsed markup.Text.
func NewMarkupText(t markup.Text) *MarkupText {
	return &MarkupText{Text: t}
}

// Render implements console.Renderable.
func (mt *MarkupText) Render(c *console.Console, opts console.Options) []segment.Segment {
	return mt.Text.Render()
}

// Measure implements console.Measurable.
func (mt *MarkupText) Measure(c *console.Console, opts console.Options) console.Measurement {
	w := cells.Len(mt.Text.Plain)
	return console.NewMeasurement(w, w)
}
