package console

import (
	"io"
	"os"
	"strings"
	"sync"

	"github.com/depado/gorich/segment"
	"github.com/depado/gorich/style"
	"golang.org/x/term"
)

// Renderable is the core protocol for anything that can be rendered.
type Renderable interface {
	Render(c *Console, opts Options) []segment.Segment
}

// Measurable is implemented by renderables that can report their width.
type Measurable interface {
	Measure(c *Console, opts Options) Measurement
}

// RenderHook allows interception of the console's print pipeline.
// Used by Live to inject cursor-repositioning codes.
type RenderHook interface {
	ProcessRenderables(renderables []Renderable) []Renderable
}

// Console handles terminal output with colors and styles.
type Console struct {
	mu          sync.Mutex
	out         io.Writer
	colorSystem style.ColorSystem
	width       int
	height      int
	isTerminal  bool
	noColor     bool
	forceTerminal *bool
	hooks       []RenderHook
}

// Option configures a Console.
type Option func(*Console)

// WithWriter sets the output writer.
func WithWriter(w io.Writer) Option {
	return func(c *Console) {
		c.out = w
	}
}

// WithColorSystem sets the color system.
func WithColorSystem(cs style.ColorSystem) Option {
	return func(c *Console) {
		c.colorSystem = cs
	}
}

// WithForceTerminal forces terminal mode on or off.
func WithForceTerminal(force bool) Option {
	return func(c *Console) {
		c.forceTerminal = &force
	}
}

// WithNoColor disables all color output.
func WithNoColor(noColor bool) Option {
	return func(c *Console) {
		c.noColor = noColor
	}
}

// WithWidth sets a fixed width (overrides terminal detection).
func WithWidth(width int) Option {
	return func(c *Console) {
		c.width = width
	}
}

// New creates a new Console.
func New(opts ...Option) *Console {
	c := &Console{
		out: os.Stdout,
	}

	for _, opt := range opts {
		opt(c)
	}

	c.detectTerminal()

	return c
}

func (c *Console) detectTerminal() {
	// Check if output is a terminal
	if c.forceTerminal != nil {
		c.isTerminal = *c.forceTerminal
	} else if f, ok := c.out.(*os.File); ok {
		c.isTerminal = term.IsTerminal(int(f.Fd()))
	}

	// Get terminal size
	if c.width == 0 || c.height == 0 {
		if f, ok := c.out.(*os.File); ok {
			if w, h, err := term.GetSize(int(f.Fd())); err == nil {
				if c.width == 0 {
					c.width = w
				}
				if c.height == 0 {
					c.height = h
				}
			}
		}
	}

	// Fallback to environment variables
	if c.width == 0 {
		if cols := os.Getenv("COLUMNS"); cols != "" {
			var w int
			if _, err := parseEnvInt(cols); err == nil {
				c.width = w
			}
		}
	}
	if c.height == 0 {
		if lines := os.Getenv("LINES"); lines != "" {
			var h int
			if _, err := parseEnvInt(lines); err == nil {
				c.height = h
			}
		}
	}

	// Final fallback
	if c.width == 0 {
		c.width = 80
	}
	if c.height == 0 {
		c.height = 25
	}

	// Detect color system
	if c.colorSystem == 0 {
		c.colorSystem = c.detectColorSystem()
	}

	// Check NO_COLOR environment variable
	if os.Getenv("NO_COLOR") != "" {
		c.noColor = true
	}
}

func parseEnvInt(s string) (int, error) {
	var n int
	_, err := strings.NewReader(s).Read([]byte{})
	return n, err
}

func (c *Console) detectColorSystem() style.ColorSystem {
	if !c.isTerminal {
		return style.ColorSystemNone
	}

	// Check COLORTERM for truecolor support
	colorTerm := os.Getenv("COLORTERM")
	if colorTerm == "truecolor" || colorTerm == "24bit" {
		return style.ColorSystemTrueColor
	}

	// Check TERM for color support
	termEnv := os.Getenv("TERM")
	if strings.Contains(termEnv, "256color") {
		return style.ColorSystem256
	}
	if strings.Contains(termEnv, "color") || strings.Contains(termEnv, "xterm") {
		return style.ColorSystemStandard
	}

	// Check for dumb terminal
	if termEnv == "dumb" || termEnv == "" {
		return style.ColorSystemNone
	}

	// Default to standard colors for unknown terminals
	return style.ColorSystemStandard
}

// Width returns the console width in cells.
func (c *Console) Width() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.width
}

// Height returns the console height in cells.
func (c *Console) Height() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.height
}

// Size returns the console dimensions.
func (c *Console) Size() Dimensions {
	c.mu.Lock()
	defer c.mu.Unlock()
	return Dimensions{Width: c.width, Height: c.height}
}

// IsTerminal returns true if output is a terminal.
func (c *Console) IsTerminal() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.isTerminal
}

// ColorSystem returns the detected color system.
func (c *Console) ColorSystem() style.ColorSystem {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.noColor {
		return style.ColorSystemNone
	}
	return c.colorSystem
}

// Options returns the default rendering options for this console.
func (c *Console) Options() Options {
	c.mu.Lock()
	defer c.mu.Unlock()
	return Options{
		Size:        Dimensions{Width: c.width, Height: c.height},
		MaxWidth:    c.width,
		IsTerminal:  c.isTerminal,
		ColorSystem: c.colorSystem,
	}
}

// PushRenderHook adds a render hook.
func (c *Console) PushRenderHook(hook RenderHook) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.hooks = append(c.hooks, hook)
}

// PopRenderHook removes the most recently added render hook.
func (c *Console) PopRenderHook() RenderHook {
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.hooks) == 0 {
		return nil
	}
	hook := c.hooks[len(c.hooks)-1]
	c.hooks = c.hooks[:len(c.hooks)-1]
	return hook
}

// Render renders a Renderable to the console.
func (c *Console) Render(r Renderable) {
	c.mu.Lock()
	defer c.mu.Unlock()

	opts := Options{
		Size:        Dimensions{Width: c.width, Height: c.height},
		MaxWidth:    c.width,
		IsTerminal:  c.isTerminal,
		ColorSystem: c.colorSystem,
	}

	// Apply render hooks
	renderables := []Renderable{r}
	for _, hook := range c.hooks {
		renderables = hook.ProcessRenderables(renderables)
	}

	// Render all renderables
	var allSegments []segment.Segment
	for _, renderable := range renderables {
		segments := renderable.Render(c, opts)
		allSegments = append(allSegments, segments...)
	}

	// Convert to ANSI string
	colorSys := c.colorSystem
	if c.noColor {
		colorSys = style.ColorSystemNone
	}

	var output strings.Builder
	for _, seg := range allSegments {
		output.WriteString(seg.Render(colorSys))
	}
	output.WriteString("\n")

	c.out.Write([]byte(output.String())) //nolint:errcheck // terminal output is fire-and-forget
}

// PrintSegments writes segments directly to the console.
func (c *Console) PrintSegments(segments []segment.Segment) {
	c.mu.Lock()
	defer c.mu.Unlock()

	colorSys := c.colorSystem
	if c.noColor {
		colorSys = style.ColorSystemNone
	}

	var output strings.Builder
	for _, seg := range segments {
		output.WriteString(seg.Render(colorSys))
	}
	output.WriteString("\n")

	c.out.Write([]byte(output.String())) //nolint:errcheck // terminal output is fire-and-forget
}

// Write writes raw bytes to the console (implements io.Writer).
func (c *Console) Write(p []byte) (n int, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.out.Write(p)
}

// WriteString writes a raw string to the console.
func (c *Console) WriteString(s string) (n int, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	n, err = c.out.Write([]byte(s))
	// Flush if the writer supports it (for flicker-free output)
	if f, ok := c.out.(interface{ Sync() error }); ok {
		f.Sync() //nolint:errcheck // best-effort flush
	}
	return n, err
}

// WriteControl writes a control sequence to the console.
func (c *Console) WriteControl(ctrl segment.Control) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.isTerminal {
		c.out.Write([]byte(ctrl.Render())) //nolint:errcheck // terminal control output is fire-and-forget
	}
}

// Text is a simple Renderable that wraps a styled string.
type Text struct {
	Content string
	Style   *style.Style
}

// NewText creates a new Text renderable.
func NewText(content string, s *style.Style) Text {
	return Text{Content: content, Style: s}
}

// Render implements Renderable.
func (t Text) Render(c *Console, opts Options) []segment.Segment {
	return []segment.Segment{segment.NewText(t.Content, t.Style)}
}

// Measure implements Measurable.
func (t Text) Measure(c *Console, opts Options) Measurement {
	// Simple measurement - count cells
	lines := strings.Split(t.Content, "\n")
	maxWidth := 0
	for _, line := range lines {
		w := len(line) // TODO: use cell width
		if w > maxWidth {
			maxWidth = w
		}
	}
	return NewMeasurement(maxWidth, maxWidth)
}

// Segments is a Renderable that wraps pre-built segments.
type Segments []segment.Segment

// Render implements Renderable.
func (s Segments) Render(c *Console, opts Options) []segment.Segment {
	return s
}

// ControlRenderable wraps a Control as a Renderable.
type ControlRenderable struct {
	Control segment.Control
}

// Render implements Renderable.
func (cr ControlRenderable) Render(c *Console, opts Options) []segment.Segment {
	return []segment.Segment{segment.NewControlSegment(cr.Control.Codes...)}
}
