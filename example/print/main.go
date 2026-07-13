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
	gorich.Println("[bold]Bold text[/]")
	gorich.Println("[italic]Italic text[/]")
	gorich.Println("[underline]Underlined text[/]")
	gorich.Println("[strike]Strikethrough text[/]")

	gorich.Rule("Colors")

	// Colors
	gorich.Println("[red]Red[/] [green]Green[/] [blue]Blue[/] [yellow]Yellow[/]")
	gorich.Println("[bright_red]Bright Red[/] [bright_green]Bright Green[/]")
	gorich.Println("[#ff6600]Hex color (orange)[/]")
	gorich.Println("[rgb(100,150,200)]RGB color[/]")

	gorich.Rule("Combined Styles")

	// Combined styles
	gorich.Println("[bold red]Bold Red[/]")
	gorich.Println("[italic green]Italic Green[/]")
	gorich.Println("[bold italic underline blue]All the styles![/]")
	gorich.Println("[white on red]White on Red background[/]")
	gorich.Println("[black on bright_yellow]Black on Bright Yellow[/]")

	gorich.Rule("Nested Markup")

	// Nested markup
	gorich.Println("[bold]This is bold and [italic]this is bold italic[/italic] back to bold[/bold]")
	gorich.Println("Normal [red]red [bold]bold red[/bold] red[/red] normal")

	gorich.Rule("Printf")

	// Printf style (no trailing newline, like standard printf)
	gorich.Printf("[bold]Count:[/] %d", 42)
	gorich.Printf(" - [green]done[/]\n")
	gorich.Printf("[cyan]Progress:[/] %.1f%%\n", 75.5)

	gorich.Rule("Escaped Markup")

	// Escaped brackets
	gorich.Println("Use \\[bold] to write literal brackets")
	gorich.Println("Array syntax: arr\\[0] = value")

	gorich.Rule("Styled Rules", console.WithRuleStyle("blue"))

	// Styled rules
	gorich.Rule("Blue Rule Line", console.WithRuleStyle("blue"))
	gorich.Rule("Bold Red Title", console.WithTitleStyle("bold red"))
	gorich.Rule("Both Styled", console.WithRuleStyle("dim"), console.WithTitleStyle("bold yellow"))
	gorich.Rule("[green]Markup[/] in Title", console.WithRuleStyle("cyan"))

	gorich.Rule("")
}
