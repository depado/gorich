package gorich

import (
	"strings"
	"testing"

	"github.com/depado/gorich/console"
	"github.com/depado/gorich/style"
)

func TestSprintf(t *testing.T) {
	c := console.New(console.WithNoColor(true), console.WithForceTerminal(true))

	if got := c.Sprintf("[bold]Count:[/] %d", 42); got != "Count: 42" {
		t.Errorf("Sprintf no-color: got %q, want %q", got, "Count: 42")
	}
	if got := c.Sprint("[bold]Hello[/]", "World"); got != "Hello World" {
		t.Errorf("Sprint no-color: got %q, want %q", got, "Hello World")
	}

	cc := console.New(console.WithForceTerminal(true), console.WithColorSystem(style.ColorSystemStandard))
	got := cc.Sprintf("[bold]Hi[/]")
	if !strings.Contains(got, "Hi") || !strings.Contains(got, "\x1b[") {
		t.Errorf("Sprintf color: expected ANSI codes around text, got %q", got)
	}
}
