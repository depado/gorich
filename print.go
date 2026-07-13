// Package gorich provides Rich-style terminal formatting for Go.
//
// Quick start:
//
//	gorich.Println("[bold red]Hello[/] [green]World[/]")
//
// For more control, create a Console:
//
//	c := console.New()
//	c.Println("[bold]Hello[/]")
package gorich

import (
	"github.com/depado/gorich/console"
)

// defaultConsole is the shared console for package-level functions.
var defaultConsole = console.New()

// Print prints Rich-style markup to stdout without a trailing newline.
//
// Example:
//
//	gorich.Print("Working... ")
//	gorich.Print("[green]done[/]\n")
func Print(args ...any) {
	defaultConsole.Print(args...)
}

// Println prints Rich-style markup to stdout with a trailing newline.
//
// Example:
//
//	gorich.Println("[bold]Hello[/] [red]World[/]")
//	gorich.Println("[italic green]Success![/]")
//	gorich.Println("[#ff0000]Hex color[/]")
//	gorich.Println("[bold red on white]Styled text[/]")
func Println(args ...any) {
	defaultConsole.Println(args...)
}

// Printf prints formatted Rich-style markup to stdout.
//
// Example:
//
//	gorich.Printf("[bold]Count:[/] %d", 42)
func Printf(format string, args ...any) {
	defaultConsole.Printf(format, args...)
}

// Sprint renders Rich-style markup args and returns the resulting ANSI string.
func Sprint(args ...any) string {
	return defaultConsole.Sprint(args...)
}

// Sprintf renders formatted Rich-style markup and returns the resulting ANSI string.
//
// Example:
//
//	s := gorich.Sprintf("[bold]Count:[/] %d", 42)
func Sprintf(format string, args ...any) string {
	return defaultConsole.Sprintf(format, args...)
}

// Log prints with a log prefix.
func Log(args ...any) {
	defaultConsole.Log(args...)
}

// Rule prints a horizontal rule with optional title.
func Rule(title string, opts ...console.RuleOption) {
	defaultConsole.Rule(title, opts...)
}

// Console returns the default console for advanced usage.
func Console() *console.Console {
	return defaultConsole
}
