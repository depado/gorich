// Command showcase combines every GoRich feature in a single demo:
// rich print & markup, rules, tables, progress bars, spinners and live blocks.
package main

import (
	"context"
	"fmt"
	"os/exec"
	"sync"
	"time"

	"github.com/depado/gorich"
	"github.com/depado/gorich/console"
	"github.com/depado/gorich/live"
	"github.com/depado/gorich/progress"
	"github.com/depado/gorich/style"
	"github.com/depado/gorich/table"
	"github.com/depado/gorich/table/align"
	"github.com/depado/gorich/table/box"
)

func main() {
	printDemo()
	tableDemo()
	progressDemo()
	spinnerDemo()
	blocksDemo()

	gorich.Rule("[bold green]All features shown[/]", console.WithRuleStyle("green"))
}

// printDemo shows markup, colors, nested styles and styled rules.
func printDemo() {
	gorich.Rule("[bold]Rich Print & Markup[/]", console.WithRuleStyle("cyan"))

	gorich.Println("[bold]bold[/] [italic]italic[/] [underline]underline[/] [strike]strike[/]")
	gorich.Println("[red]red[/] [green]green[/] [blue]blue[/] [#ff6600]hex[/] [rgb(100,150,200)]rgb[/]")
	gorich.Println("[bold red on white]bold red on white[/]")
	gorich.Println("Normal [red]red [bold]bold red[/bold] red[/red] normal")
	gorich.Printf("[cyan]Progress:[/] %.1f%%\n", 75.5)
	gorich.Println()
}

// tableDemo shows columns, styling, footers, sections, vertical alignment,
// expand + ratios and an alternate box style.
func tableDemo() {
	c := gorich.Console()
	gorich.Rule("[bold]Tables[/]", console.WithRuleStyle("cyan"))

	tbl := table.NewTableWithOptions(
		nil,
		table.WithTitle("[bold underline]Sales Report[/]"),
		table.WithBox(box.ROUNDED),
		table.WithShowFooter(true),
		table.WithHeaderStyle(new(style.Parse("bold magenta"))),
		table.WithFooterStyle(new(style.Parse("bold cyan"))),
		table.WithRowStyles(
			new(style.Parse("")),
			new(style.Parse("on #222222")),
		),
	)
	colProduct := tbl.AddColumn("Product", table.WithColumnStyle("bold"))
	colQty := tbl.AddColumn("Qty", table.WithColumnJustify(console.JustifyCenter))
	colTotal := tbl.AddColumn("Total", table.WithColumnJustify(console.JustifyRight), table.WithColumnStyle("bold green"))
	colProduct.Footer = "3 products"
	colQty.Footer = "64"
	colTotal.Footer = "[bold]$291.43[/]"
	tbl.AddRow("Widget A", "15", "$67.50")
	tbl.AddRow("[blue]Widget B[/]", "[bold]42[/]", "$84.00")
	tbl.AddSection()
	tbl.AddRow("[italic]Gadget X[/]", "7", "$139.93")
	c.Render(tbl)
	c.Println()

	// Vertical alignment + expand/ratio columns.
	tbl2 := table.NewTableWithOptions(nil, table.WithExpand(), table.WithTitle("[bold]Expand + Vertical[/]"))
	colLabel := tbl2.AddColumn("Label", table.WithColumnWidth(10), table.WithColumnStyle("bold cyan"))
	colBody := tbl2.AddColumn("Content", table.WithColumnRatio(1))
	colLabel.Vertical = align.Middle
	colBody.Vertical = align.Middle
	tbl2.AddRow("Multi", "[dim]This text\nspans multiple lines\nand is centered vertically[/]")
	tbl2.AddRow("Single", "Fills the flexible column and wraps to fit the width")
	c.Render(tbl2)
	c.Println()
}

// progressDemo shows multiple concurrent tasks with download-style columns.
func progressDemo() {
	gorich.Rule("[bold]Progress Bars[/]", console.WithRuleStyle("cyan"))

	p := progress.New(
		progress.WithRefreshRate(30),
		progress.WithColumns(
			progress.NewSpinnerColumn(),
			progress.DescriptionColumn(),
			progress.NewBarColumn(progress.WithBarWidth(30)),
			progress.NewTaskProgressColumn(false),
			progress.NewSeparatorColumn("•"),
			progress.NewTimeRemainingColumn(),
		),
	)
	ctx := context.Background()
	p.Start(ctx)

	t1, t2, t3 := 50.0, 100.0, 75.0
	id1 := p.AddTask("[cyan]Downloading[/]", &t1)
	id2 := p.AddTask("[yellow]Processing[/]", &t2)
	id3 := p.AddTask("[magenta]Uploading[/]", &t3)

	for i := range 100 {
		time.Sleep(25 * time.Millisecond)
		if i < 50 {
			p.Advance(id1, 1)
		}
		p.Advance(id2, 1)
		if i < 75 {
			p.Advance(id3, 1)
		}
	}
	p.Stop()
	fmt.Println()
}

// spinnerDemo runs several concurrent tasks each with a distinct spinner.
func spinnerDemo() {
	gorich.Rule("[bold]Spinners[/]", console.WithRuleStyle("cyan"))

	spinnerCol := progress.NewSpinnerColumn()
	p := progress.New(
		progress.WithRefreshRate(20),
		progress.WithColumns(
			spinnerCol,
			progress.DescriptionColumn(),
			progress.NewBarColumn(progress.WithBarWidth(25)),
			progress.NewMofNCompleteColumn(" / "),
		),
	)
	ctx := context.Background()
	p.Start(ctx)

	tasks := []struct {
		label   string
		spinner string
		total   float64
		delay   time.Duration
	}{
		{"[green]Compiling[/]", "dots", 80, 25 * time.Millisecond},
		{"[yellow]Linting[/]", "arc", 60, 35 * time.Millisecond},
		{"[magenta]Testing[/]", "triangle", 100, 20 * time.Millisecond},
	}

	ids := make([]progress.TaskID, len(tasks))
	for i, t := range tasks {
		total := t.total
		ids[i] = p.AddTask(t.label, &total)
		spinnerCol.SetTaskSpinner(ids[i], t.spinner)
	}

	var wg sync.WaitGroup
	for i, t := range tasks {
		wg.Add(1)
		go func(id progress.TaskID, total float64, delay time.Duration) {
			defer wg.Done()
			for j := 0.0; j < total; j++ {
				time.Sleep(delay)
				p.Advance(id, 1)
			}
		}(ids[i], t.total, t.delay)
	}
	wg.Wait()
	p.Stop()
	fmt.Println()
}

// blocksDemo streams parallel command output into per-task live blocks.
func blocksDemo() {
	gorich.Rule("[bold]Live Blocks[/]", console.WithRuleStyle("cyan"))

	c := console.New()
	display := live.NewBlockDisplay(
		live.WithBlockMaxLines(3),
		live.WithBlockSpinnerName("dots"),
		live.WithBlockPrefix("[dim]│ [/]"),
	)
	l := live.New(c, display, live.WithAutoRefresh(true), live.WithRefreshRate(15))
	ctx := context.Background()
	l.Start(ctx)

	repos := []struct{ name, expr string }{
		{"alpha", "echo cloning && sleep 1 && echo done"},
		{"beta", "echo start && sleep 1 && echo middle && sleep 0.5 && echo more"},
		{"gamma", "echo failing && sleep 1 && exit 2"},
	}

	var wg sync.WaitGroup
	for _, r := range repos {
		wg.Add(1)
		go func(name, expr string) {
			defer wg.Done()
			idx := display.Start(name)
			out := display.NewWriter(idx)
			cmd := exec.CommandContext(ctx, "sh", "-c", expr)
			cmd.Stdout = out
			cmd.Stderr = out
			if err := cmd.Run(); err != nil {
				if exitErr, ok := err.(*exec.ExitError); ok {
					display.Finish(idx, exitErr.ExitCode())
				} else {
					display.Finish(idx, 1)
				}
				return
			}
			display.Finish(idx, 0)
		}(r.name, r.expr)
	}
	wg.Wait()
	time.Sleep(300 * time.Millisecond)
	l.Stop()
}
