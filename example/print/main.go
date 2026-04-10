package main

import (
	"os"

	"github.com/depado/gorich"
	"github.com/depado/gorich/console"
)

func main() {
	// Force color output for demo (normally auto-detected)
	if os.Getenv("FORCE_COLOR") != "" {
		c := console.New(console.WithForceTerminal(true))
		// Use this console instead of the default
		_ = c
	}
	gorich.Rule("Rich Print Demo")

	// Basic markup
	gorich.Print("[bold]Bold text[/]")
	gorich.Print("[italic]Italic text[/]")
	gorich.Print("[underline]Underlined text[/]")
	gorich.Print("[strike]Strikethrough text[/]")

	gorich.Rule("Colors")

	// Colors
	gorich.Print("[red]Red[/] [green]Green[/] [blue]Blue[/] [yellow]Yellow[/]")
	gorich.Print("[bright_red]Bright Red[/] [bright_green]Bright Green[/]")
	gorich.Print("[#ff6600]Hex color (orange)[/]")
	gorich.Print("[rgb(100,150,200)]RGB color[/]")

	gorich.Rule("Combined Styles")

	// Combined styles
	gorich.Print("[bold red]Bold Red[/]")
	gorich.Print("[italic green]Italic Green[/]")
	gorich.Print("[bold italic underline blue]All the styles![/]")
	gorich.Print("[white on red]White on Red background[/]")
	gorich.Print("[black on bright_yellow]Black on Bright Yellow[/]")

	gorich.Rule("Nested Markup")

	// Nested markup
	gorich.Print("[bold]This is bold and [italic]this is bold italic[/italic] back to bold[/bold]")
	gorich.Print("Normal [red]red [bold]bold red[/bold] red[/red] normal")

	gorich.Rule("Printf")

	// Printf style
	gorich.Printf("[bold]Count:[/] %d", 42)
	gorich.Printf("[green]Status:[/] %s", "OK")
	gorich.Printf("[cyan]Progress:[/] %.1f%%", 75.5)

	gorich.Rule("Escaped Markup")

	// Escaped brackets
	gorich.Print("Use \\[bold] to write literal brackets")
	gorich.Print("Array syntax: arr\\[0] = value")

	gorich.Rule("")
}
