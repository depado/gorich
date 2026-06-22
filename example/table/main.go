package main

import (
	"github.com/depado/gorich"
	"github.com/depado/gorich/console"
	"github.com/depado/gorich/style"
	"github.com/depado/gorich/table"
	"github.com/depado/gorich/table/align"
	"github.com/depado/gorich/table/box"
)

func main() {
	c := gorich.Console()

	// ── Column styling & markup ──
	c.Printf("[bold]Column styling + markup in cells:[/]\n")
	tbl := table.NewTable()
	tbl.AddColumn("Name", table.WithColumnStyle("bold"))
	tbl.AddColumn("Age", table.WithColumnJustify(console.JustifyCenter), table.WithColumnStyle("yellow"))
	tbl.AddColumn("City", table.WithColumnStyle("italic"))
	tbl.AddRow("[bold]Alice[/]", "30", "New York")
	tbl.AddRow("Bob", "25", "[blue]San Francisco[/]")
	tbl.AddRow("[red]Charlie[/]", "35", "[green]London[/]")
	c.Render(tbl)
	c.Print()

	// ── Right-aligned numbers, alternating rows ──
	c.Printf("[bold]Sales report with right-aligned numbers, alternating rows:[/]\n")
	tbl2 := table.NewTableWithOptions(
		nil,
		table.WithTitle("[bold underline]Sales Report[/]"),
		table.WithRowStyles(
			stylePtr(style.Parse("")),
			stylePtr(style.Parse("on #222222")),
		),
	)
	tbl2.AddColumn("Product")
	tbl2.AddColumn("Qty", table.WithColumnJustify(console.JustifyCenter))
	tbl2.AddColumn("Price", table.WithColumnJustify(console.JustifyRight), table.WithColumnStyle("yellow"))
	tbl2.AddColumn("Total", table.WithColumnJustify(console.JustifyRight), table.WithColumnStyle("bold green"))
	tbl2.AddRow("Widget A", "15", "$4.50", "[bold]$67.50[/]")
	tbl2.AddRow("Widget B", "[bold]42[/]", "$2.00", "$84.00")
	tbl2.AddRow("Gadget X", "7", "$19.99", "$139.93")
	tbl2.AddRow("[italic]Gadget Y[/]", "3", "$29.99", "[bold]$89.97[/]")
	c.Render(tbl2)
	c.Print()

	// ── Footer + sections ──
	c.Printf("[bold]Footer + sections + show_lines:[/]\n")
	tbl3 := table.NewTableWithOptions(
		nil,
		table.WithShowFooter(true),
		table.WithHeaderStyle(stylePtr(style.Parse("bold magenta"))),
		table.WithFooterStyle(stylePtr(style.Parse("bold cyan"))),
		table.WithShowLines(),
	)
	colTask := tbl3.AddColumn("Task", table.WithColumnStyle("bold"))
	colStatus := tbl3.AddColumn("Status", table.WithColumnJustify(console.JustifyCenter))
	colAssignee := tbl3.AddColumn("Assignee")
	colTask.Footer = "Total: 5 tasks"
	colStatus.Footer = "[bold]3 done[/]"
	colAssignee.Footer = ""

	tbl3.AddRow("[bold]Setup project[/]", "[green]Done[/]", "Alice")
	tbl3.AddRow("[bold]Write tests[/]", "[green]Done[/]", "Bob")
	tbl3.AddRow("[bold]Implement feature[/]", "[yellow]In Progress[/]", "Charlie")
	tbl3.AddSection()
	tbl3.AddRow("[bold]Code review[/]", "[dim]Pending[/]", "Alice")
	tbl3.AddRow("[bold]Deploy[/]", "[dim]Pending[/]", "Bob")
	c.Render(tbl3)
	c.Print()

	// ── Vertical alignment (top) ──
	c.Printf("[bold]Vertical alignment (top):[/]\n")
	tbl4a := table.NewTableWithOptions(
		nil,
		table.WithTitle("[bold]Vertical: Top[/]"),
	)
	colLabelA := tbl4a.AddColumn("Label", table.WithColumnWidth(10), table.WithColumnStyle("bold cyan"))
	colDescA := tbl4a.AddColumn("Content", table.WithColumnWidth(32))
	colLabelA.Vertical = align.Top
	colDescA.Vertical = align.Top
	tbl4a.AddRow("Multi", "[dim]This text\nspans multiple lines\nand is top-aligned[/]")
	tbl4a.AddRow("Aligned", "Single-line text sits at the top")
	c.Render(tbl4a)
	c.Print()

	// ── Vertical alignment (middle) ──
	c.Printf("[bold]Vertical alignment (middle):[/]\n")
	tbl4b := table.NewTableWithOptions(
		nil,
		table.WithTitle("[bold]Vertical: Middle[/]"),
	)
	colLabelB := tbl4b.AddColumn("Label", table.WithColumnWidth(10), table.WithColumnStyle("bold cyan"))
	colDescB := tbl4b.AddColumn("Content", table.WithColumnWidth(32))
	colLabelB.Vertical = align.Middle
	colDescB.Vertical = align.Middle
	tbl4b.AddRow("Multi", "[dim]This text\nspans multiple lines\nand is middle-aligned[/]")
	tbl4b.AddRow("Aligned", "Single-line text is centered vertically")
	c.Render(tbl4b)
	c.Print()

	// ── Vertical alignment (bottom) ──
	c.Printf("[bold]Vertical alignment (bottom):[/]\n")
	tbl4c := table.NewTableWithOptions(
		nil,
		table.WithTitle("[bold]Vertical: Bottom[/]"),
	)
	colLabelC := tbl4c.AddColumn("Label", table.WithColumnWidth(10), table.WithColumnStyle("bold cyan"))
	colDescC := tbl4c.AddColumn("Content", table.WithColumnWidth(32))
	colLabelC.Vertical = align.Bottom
	colDescC.Vertical = align.Bottom
	tbl4c.AddRow("Multi", "[dim]This text\nspans multiple lines\nand is bottom-aligned[/]")
	tbl4c.AddRow("Aligned", "Single-line text sits at the bottom")
	c.Render(tbl4c)
	c.Print()

	// ── Expand with ratios ──
	c.Printf("[bold]Expand + ratio columns:[/]\n")
	tbl5 := table.NewTableWithOptions(
		nil,
		table.WithExpand(),
		table.WithTitle("[bold]Expand + Ratio Columns[/]"),
	)
	tbl5.AddColumn("Small", table.WithColumnWidth(12))
	tbl5.AddColumn("Flexible (2x)", table.WithColumnRatio(2), table.WithColumnStyle("cyan"), table.WithColumnNoWrap())
	tbl5.AddColumn("Fixed (1x)", table.WithColumnRatio(1), table.WithColumnStyle("yellow"), table.WithColumnNoWrap())
	tbl5.AddRow("narrow col", "This column gets 2x the space", "This gets 1x")
	tbl5.AddRow("still 12", "Longer content here will be truncated", "Shorter")
	c.Render(tbl5)
	c.Print()

	// ── All box styles gallery ──
	c.Printf("[bold]Box Style Gallery:[/]\n")
	boxStyles := []struct {
		name string
		b    *box.Box
	}{
		{"SIMPLE", box.SIMPLE},
		{"SIMPLE_HEAVY", box.SIMPLE_HEAVY},
		{"ASCII", box.ASCII},
		{"ASCII2", box.ASCII2},
		{"MINIMAL", box.MINIMAL},
		{"MINIMAL_HEAVY_HEAD", box.MINIMAL_HEAVY_HEAD},
		{"HORIZONTALS", box.HORIZONTALS},
		{"ROUNDED", box.ROUNDED},
		{"SQUARE", box.SQUARE},
		{"SQUARE_DOUBLE_HEAD", box.SQUARE_DOUBLE_HEAD},
		{"HEAVY", box.HEAVY},
		{"HEAVY_EDGE", box.HEAVY_EDGE},
		{"HEAVY_HEAD", box.HEAVY_HEAD},
		{"DOUBLE", box.DOUBLE},
		{"DOUBLE_EDGE", box.DOUBLE_EDGE},
		{"MARKDOWN", box.MARKDOWN},
	}
	for _, bs := range boxStyles {
		bt := table.NewTableWithOptions(
			nil,
			table.WithBox(bs.b),
			table.WithTitle("[bold dim]box." + bs.name + "[/]"),
		)
		bt.AddColumn("Property", table.WithColumnStyle("cyan"))
		bt.AddColumn("Value", table.WithColumnStyle("green"))
		bt.AddRow("Style", bs.name)
		bt.AddRow("Border", "[dim]rendered[/]")
		c.Render(bt)
	}
}

func stylePtr(s style.Style) *style.Style {
	return &s
}
