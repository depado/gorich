package console

import (
	"fmt"
	"strings"

	"github.com/depado/gorich/markup"
	"github.com/depado/gorich/segment"
)

// Print prints Rich-style markup to the console.
// Supports markup syntax like [bold red]Hello[/bold] World.
//
// Example:
//
//	console.Print("[bold]Hello[/] [red]World[/]")
//	console.Print("[italic green]Success![/]")
func (c *Console) Print(args ...any) {
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
		colorSys = 0
	}

	var output strings.Builder
	for _, seg := range segments {
		output.WriteString(seg.Render(colorSys))
	}
	output.WriteString("\n")

	c.out.Write([]byte(output.String())) //nolint:errcheck // terminal output is fire-and-forget
}

// Printf prints formatted Rich-style markup to the console.
//
// Example:
//
//	console.Printf("[bold]Count:[/] %d", count)
func (c *Console) Printf(format string, args ...any) {
	text := fmt.Sprintf(format, args...)
	c.Print(text)
}

// PrintMarkup prints pre-parsed markup segments.
func (c *Console) PrintMarkup(text markup.Text) {
	c.mu.Lock()
	defer c.mu.Unlock()

	segments := text.Render()

	colorSys := c.colorSystem
	if c.noColor {
		colorSys = 0
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
		colorSys = 0
	}

	var output strings.Builder

	// Add timestamp in dim style
	// TODO: add actual timestamp formatting
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
	// Simple time format - could be made configurable
	return "LOG"
}

// Rule prints a horizontal rule with optional title.
func (c *Console) Rule(title string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	width := c.width
	if width <= 0 {
		width = 80
	}

	var output strings.Builder

	if title == "" {
		// Just a line
		output.WriteString(strings.Repeat("─", width))
	} else {
		// Title centered in the rule
		titleLen := len(title) + 2 // space on each side
		if titleLen >= width-4 {
			output.WriteString(title)
		} else {
			leftLen := (width - titleLen) / 2
			rightLen := width - titleLen - leftLen
			output.WriteString(strings.Repeat("─", leftLen))
			output.WriteString(" ")
			output.WriteString(title)
			output.WriteString(" ")
			output.WriteString(strings.Repeat("─", rightLen))
		}
	}
	output.WriteString("\n")

	c.out.Write([]byte(output.String())) //nolint:errcheck // terminal output is fire-and-forget
}
