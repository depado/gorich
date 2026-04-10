// Package gorich provides Rich-style terminal formatting for Go.
//
// Quick start:
//
//	gorich.Print("[bold red]Hello[/] [green]World[/]")
//
// For more control, create a Console:
//
//	c := console.New()
//	c.Print("[bold]Hello[/]")
package gorich

import (
	"github.com/depado/gorich/console"
)

// defaultConsole is the shared console for package-level functions.
var defaultConsole = console.New()

// Print prints Rich-style markup to stdout.
//
// Example:
//
//	gorich.Print("[bold]Hello[/] [red]World[/]")
//	gorich.Print("[italic green]Success![/]")
//	gorich.Print("[#ff0000]Hex color[/]")
//	gorich.Print("[bold red on white]Styled text[/]")
func Print(args ...any) {
	defaultConsole.Print(args...)
}

// Printf prints formatted Rich-style markup to stdout.
//
// Example:
//
//	gorich.Printf("[bold]Count:[/] %d", 42)
func Printf(format string, args ...any) {
	defaultConsole.Printf(format, args...)
}

// Log prints with a log prefix.
func Log(args ...any) {
	defaultConsole.Log(args...)
}

// Rule prints a horizontal rule with optional title.
func Rule(title string) {
	defaultConsole.Rule(title)
}

// Console returns the default console for advanced usage.
func Console() *console.Console {
	return defaultConsole
}
