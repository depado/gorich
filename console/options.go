// Package console provides the terminal rendering engine.
package console

import "github.com/depado/gorich/style"

// Dimensions represents terminal dimensions in cells.
type Dimensions struct {
	Width  int
	Height int
}

// Options carries rendering constraints passed to every Renderable.
type Options struct {
	Size        Dimensions
	MinWidth    int
	MaxWidth    int
	MaxHeight   int
	IsTerminal  bool
	ASCIIOnly   bool
	ColorSystem style.ColorSystem
	NoWrap      bool
	Overflow    Overflow
	Justify     Justify
}

// Overflow determines how text overflow is handled.
type Overflow int

const (
	OverflowFold    Overflow = iota // Wrap at width
	OverflowCrop                    // Truncate at width
	OverflowEllipsis                // Truncate with ellipsis
)

// Justify determines text justification.
type Justify int

const (
	JustifyDefault Justify = iota
	JustifyLeft
	JustifyCenter
	JustifyRight
	JustifyFull
)

// WithMaxWidth returns a copy with the given max width.
func (o Options) WithMaxWidth(width int) Options {
	o.MaxWidth = width
	return o
}

// WithNoWrap returns a copy with no-wrap set.
func (o Options) WithNoWrap(noWrap bool) Options {
	o.NoWrap = noWrap
	return o
}

// Measurement represents the min and max width of a renderable.
type Measurement struct {
	Minimum int
	Maximum int
}

// NewMeasurement creates a measurement with the given min and max.
func NewMeasurement(min, max int) Measurement {
	return Measurement{Minimum: min, Maximum: max}
}

// Clamp clamps a width to this measurement's bounds.
func (m Measurement) Clamp(width int) int {
	if width < m.Minimum {
		return m.Minimum
	}
	if width > m.Maximum {
		return m.Maximum
	}
	return width
}

// Span returns max - min.
func (m Measurement) Span() int {
	return m.Maximum - m.Minimum
}
