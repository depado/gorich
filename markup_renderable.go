package gorich

import (
	"github.com/depado/gorich/console"
	"github.com/depado/gorich/markup"
	"github.com/depado/gorich/segment"
)

// MarkupRenderable wraps a string as a console.Renderable that parses
// Rich-style markup on render. It implements both console.Renderable
// and console.Measurable.
type MarkupRenderable struct {
	content string
}

// NewMarkupRenderable creates a MarkupRenderable from a markup string.
func NewMarkupRenderable(content string) *MarkupRenderable {
	return &MarkupRenderable{content: content}
}

// Render implements console.Renderable.
func (mr *MarkupRenderable) Render(c *console.Console, opts console.Options) []segment.Segment {
	return markup.Render(mr.content)
}

// Measure implements console.Measurable.
func (mr *MarkupRenderable) Measure(c *console.Console, opts console.Options) console.Measurement {
	w := markup.VisibleLength(mr.content)
	return console.NewMeasurement(w, w)
}

// Content returns the raw markup string.
func (mr *MarkupRenderable) Content() string {
	return mr.content
}
