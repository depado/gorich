package console

import (
	"fmt"
	"strings"
	"time"

	"github.com/depado/gorich/internal/cells"
	"github.com/depado/gorich/markup"
	"github.com/depado/gorich/segment"
	"github.com/depado/gorich/style"
)

// Print prints Rich-style markup to the console with a trailing newline.
// Supports markup syntax like [bold red]Hello[/bold] World.
//
// Example:
//
//	console.Print("[bold]Hello[/] [red]World[/]")
//	console.Print("[italic green]Success![/]")
func (c *Console) Print(args ...any) {
	var parts []string
	for _, arg := range args {
		parts = append(parts, fmt.Sprint(arg))
	}
	c.Printf("%s\n", strings.Join(parts, " "))
}

// Printf prints formatted Rich-style markup to the console without a trailing newline.
//
// Example:
//
//	console.Printf("[bold]Count:[/] %d\n", count)
func (c *Console) Printf(format string, args ...any) {
	c.mu.Lock()
	defer c.mu.Unlock()

	text := fmt.Sprintf(format, args...)

	// Parse markup and render
	segments := markup.Render(text)

	// Convert to ANSI string
	colorSys := c.colorSystem
	if c.noColor {
		colorSys = style.ColorSystemNone
	}

	var output strings.Builder
	for _, seg := range segments {
		output.WriteString(seg.Render(colorSys))
	}

	c.out.Write([]byte(output.String())) //nolint:errcheck // terminal output is fire-and-forget
}

// PrintMarkup prints pre-parsed markup segments.
func (c *Console) PrintMarkup(text markup.Text) {
	c.mu.Lock()
	defer c.mu.Unlock()

	segments := text.Render()

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

// Log prints with a timestamp prefix (like Rich's console.log).
func (c *Console) Log(args ...any) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Build the output string
	var parts []string
	for _, arg := range args {
		parts = append(parts, fmt.Sprint(arg))
	}
	text := strings.Join(parts, " ")

	// Parse markup and render
	segments := markup.Render(text)

	// Convert to ANSI string
	colorSys := c.colorSystem
	if c.noColor {
		colorSys = style.ColorSystemNone
	}

	var output strings.Builder

	// Add timestamp in dim style
	output.WriteString(segment.Segment{
		Text:  "[" + currentTime() + "] ",
		Style: nil, // Could add dim style here
	}.Render(colorSys))

	for _, seg := range segments {
		output.WriteString(seg.Render(colorSys))
	}
	output.WriteString("\n")

	c.out.Write([]byte(output.String())) //nolint:errcheck // terminal output is fire-and-forget
}

func currentTime() string {
	// Format like Python Rich: HH:MM:SS
	return time.Now().Format("15:04:05")
}

// RuleOption configures Rule rendering.
type RuleOption func(*ruleConfig)

type ruleConfig struct {
	style      style.Style
	titleStyle style.Style
	hasStyle   bool
	hasTitle   bool
}

// WithRuleStyle sets the style for the rule line characters.
func WithRuleStyle(s string) RuleOption {
	return func(cfg *ruleConfig) {
		cfg.style = style.Parse(s)
		cfg.hasStyle = true
	}
}

// WithTitleStyle sets the style for the title text.
func WithTitleStyle(s string) RuleOption {
	return func(cfg *ruleConfig) {
		cfg.titleStyle = style.Parse(s)
		cfg.hasTitle = true
	}
}

// Rule prints a horizontal rule with optional title.
func (c *Console) Rule(title string, opts ...RuleOption) {
	c.mu.Lock()
	defer c.mu.Unlock()

	cfg := &ruleConfig{}
	for _, opt := range opts {
		opt(cfg)
	}

	width := c.width
	if width <= 0 {
		width = 80
	}

	colorSys := c.colorSystem
	if c.noColor {
		colorSys = style.ColorSystemNone
	}

	var output strings.Builder

	// Helper to get style pointer for segments
	var ruleStyle *style.Style
	if cfg.hasStyle {
		ruleStyle = &cfg.style
	}

	if title == "" {
		seg := segment.Segment{Text: strings.Repeat("─", width), Style: ruleStyle}
		output.WriteString(seg.Render(colorSys))
	} else {
		// Parse markup in title and get visible length
		titleSegments := markup.Render(title)
		titleLen := 0
		for _, seg := range titleSegments {
			titleLen += cells.Len(seg.Text)
		}
		titleLen += 2 // space on each side

		if titleLen >= width-4 {
			// Title too long, just render it
			for _, seg := range titleSegments {
				if cfg.hasTitle {
					combined := cfg.titleStyle
					if seg.Style != nil {
						combined = cfg.titleStyle.Add(*seg.Style)
					}
					seg.Style = &combined
				}
				output.WriteString(seg.Render(colorSys))
			}
		} else {
			leftLen := (width - titleLen) / 2
			rightLen := width - titleLen - leftLen

			// Left rule
			leftSeg := segment.Segment{Text: strings.Repeat("─", leftLen), Style: ruleStyle}
			output.WriteString(leftSeg.Render(colorSys))

			// Space + title + space
			output.WriteString(" ")
			for _, seg := range titleSegments {
				if cfg.hasTitle {
					combined := cfg.titleStyle
					if seg.Style != nil {
						combined = cfg.titleStyle.Add(*seg.Style)
					}
					seg.Style = &combined
				}
				output.WriteString(seg.Render(colorSys))
			}
			output.WriteString(" ")

			// Right rule
			rightSeg := segment.Segment{Text: strings.Repeat("─", rightLen), Style: ruleStyle}
			output.WriteString(rightSeg.Render(colorSys))
		}
	}
	output.WriteString("\n")

	c.out.Write([]byte(output.String())) //nolint:errcheck // terminal output is fire-and-forget
}
