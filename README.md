# GoRich

A Go port of Python's [Rich](https://github.com/Textualize/rich) library for beautiful terminal output. Features styled text with markup syntax, progress bars, tables, and customizable displays.

## Features

- **Rich Print** - Styled text with `[bold red]markup[/]` syntax
- **Progress Bars** - Multiple concurrent tasks with customizable columns
- **Tables** - Bordered tables with column styling, markup cells, row styles, footers, sections, and 19 box styles
- **Flicker-free** - Single-write buffered output for smooth updates
- **Speed estimation** - ETA calculation with rolling average
- **File progress** - `io.Reader`/`io.Writer` wrappers for IO tracking
- **50+ spinners** - Animated spinners from cli-spinners
- **Color support** - Truecolor, 256-color, and 16-color with auto-downgrading
- **Thread-safe** - Safe for concurrent updates from multiple goroutines

## Installation

```bash
go get github.com/depado/gorich
```

## Quick Start - Rich Print

```go
package main

import (
    "github.com/depado/gorich"
    "github.com/depado/gorich/console"
)

func main() {
    // Styled text with markup
    gorich.Print("[bold]Hello[/] [red]World[/]")
    gorich.Print("[italic green]Success![/]")
    gorich.Print("[bold white on blue]Highlighted[/]")
    
    // Hex and RGB colors
    gorich.Print("[#ff6600]Orange[/]")
    gorich.Print("[rgb(100,150,200)]Custom color[/]")
    
    // Printf style (no trailing newline, like standard printf)
    gorich.Printf("[bold]Count:[/] %d", 42)
    gorich.Printf(" - [green]done[/]\n")
    
    // Horizontal rules with optional styling
    gorich.Rule("Section Title")
    gorich.Rule("Styled Rule", console.WithRuleStyle("blue"))
    gorich.Rule("Styled Title", console.WithTitleStyle("bold red"))
}
```

### Markup Syntax

| Syntax | Description |
|--------|-------------|
| `[bold]text[/]` | Bold text |
| `[italic]text[/]` | Italic text* |
| `[underline]text[/]` | Underlined text |
| `[strike]text[/]` | Strikethrough |
| `[red]text[/]` | Named colors |
| `[bright_red]text[/]` | Bright variants |
| `[#ff0000]text[/]` | Hex colors |
| `[rgb(255,0,0)]text[/]` | RGB colors |
| `[bold red]text[/]` | Combined styles |
| `[white on red]text[/]` | Background colors |
| `[bold red on white]text[/]` | Full style |
| `\\[text]` | Escaped brackets |

*Italic support depends on your terminal and font. Many terminals don't support italic or require a font with italic glyphs.

### Horizontal Rules

```go
// Simple rule
gorich.Rule("Section Title")

// Style the rule line
gorich.Rule("Blue Line", console.WithRuleStyle("blue"))

// Style the title
gorich.Rule("Bold Title", console.WithTitleStyle("bold red"))

// Style both
gorich.Rule("Fancy", console.WithRuleStyle("dim"), console.WithTitleStyle("bold yellow"))

// Markup in title (combined with title style)
gorich.Rule("[green]Success[/]", console.WithRuleStyle("cyan"))

// Empty rule (just a line)
gorich.Rule("")
```

## Progress Bars

```go
package main

import (
    "context"
    "time"

    "github.com/depado/gorich/progress"
)

func main() {
    p := progress.New()
    p.Start(context.Background())
    defer p.Stop()

    total := 100.0
    task := p.AddTask("Processing", &total)

    for i := 0; i < 100; i++ {
        time.Sleep(50 * time.Millisecond)
        p.Advance(task, 1)
    }
}
```

Output:
```
  Processing ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━  75% • 0:00:02
```

## Multiple Tasks

Task descriptions support the same markup syntax as `gorich.Print()`:

```go
p := progress.New()
p.Start(context.Background())
defer p.Stop()

total1, total2, total3 := 100.0, 200.0, 150.0

// Descriptions support [markup] syntax for colors and styles
task1 := p.AddTask("[cyan]Downloading[/]", &total1)
task2 := p.AddTask("[yellow]Processing[/]", &total2)
task3 := p.AddTask("[magenta]Uploading[/]", &total3)

// Update tasks concurrently - it's thread-safe
go func() {
    for i := 0; i < 100; i++ {
        p.Advance(task1, 1)
        time.Sleep(20 * time.Millisecond)
    }
}()
// ... similar for other tasks
```

Output:
```
 Downloading ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━ 100% • 0:00:00
  Processing ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━  65% • 0:00:03
   Uploading ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━  80% • 0:00:01
```

## Custom Columns

Customize the progress display with different column types:

```go
p := progress.New(
    progress.WithColumns(
        progress.NewSpinnerColumn(),
        progress.DescriptionColumn(),
        progress.NewBarColumn(progress.WithBarWidth(30)),
        progress.NewDownloadColumn(false),
        progress.NewTransferSpeedColumn(false),
        progress.NewTimeRemainingColumn(),
    ),
)
```

### Available Columns

| Column | Description | Example Output |
|--------|-------------|----------------|
| `DescriptionColumn()` | Task description (supports markup) | `Downloading` |
| `NewBarColumn()` | Visual progress bar (turns green when done) | `━━━━━━━━━━━━━━━━` |
| `NewTaskProgressColumn(showSpeed)` | Percentage or speed | `75%` |
| `NewTimeRemainingColumn()` | Estimated time remaining | `0:00:15` |
| `NewTimeElapsedColumn()` | Elapsed time | `0:01:23` |
| `NewSpinnerColumn()` | Animated spinner | `⠋` |
| `NewDownloadColumn(binary)` | Download progress | `5.2 MB/10.0 MB` |
| `NewTransferSpeedColumn(binary)` | Transfer rate | `1.5 MB/s` |
| `NewMofNCompleteColumn(sep)` | M of N items | `50/100` |
| `NewSeparatorColumn(text)` | Static separator | `•` |

## File Progress

Track progress while reading files:

```go
p := progress.New()
p.Start(context.Background())
defer p.Stop()

// Wrap a file with progress tracking
reader, taskID, err := p.WrapFile("large-file.bin", "Reading")
if err != nil {
    log.Fatal(err)
}
defer reader.Close()

// Read as normal - progress updates automatically
io.Copy(io.Discard, reader)
```

Or wrap any `io.Reader`:

```go
resp, _ := http.Get("https://example.com/file.zip")
reader, taskID := p.WrapReader(resp.Body, resp.ContentLength, "Downloading")
io.Copy(file, reader)
```

## Indeterminate Progress

For tasks with unknown total, pass `nil` - the bar will show a pulsing animation:

```go
task := p.AddTask("Searching", nil)  // nil total = indeterminate

// Later, when you know the total:
total := 500.0
p.Update(task, progress.TaskUpdateConfig{Total: &total})
```

## Configuration Options

```go
p := progress.New(
    progress.WithConsole(customConsole),       // Custom console
    progress.WithRefreshRate(15),              // 15 Hz refresh (default: 10)
    progress.WithSpeedEstimatePeriod(60),      // 60s speed window (default: 30)
    progress.WithTransient(true),              // Clear display when done
    progress.WithDisable(true),                // Disable output (for CI)
)
```

## Task Options

```go
task := p.AddTask("Processing", &total,
    progress.TaskWithCompleted(50),           // Start at 50%
    progress.TaskWithVisible(false),          // Hidden initially
    progress.TaskWithStart(false),            // Don't start timer yet
    progress.TaskWithFields(map[string]any{   // Custom fields
        "filename": "data.csv",
    }),
)

// Control task timing
p.StartTask(task)   // Start the timer
p.StopTask(task)    // Pause the timer
p.ResetTask(task, true)  // Reset and restart
```

## Tables

Create bordered tables with styled columns, markup cells, row formatting, footers, sections, flexible column widths, and 19 box styles.

```go
package main

import (
    "github.com/depado/gorich/console"
    "github.com/depado/gorich/table"
)

func main() {
    c := console.New()

    tbl := table.NewTable("Name", "Age", "City")
    tbl.AddRow("Alice", "30", "New York")
    tbl.AddRow("Bob", "25", "San Francisco")
    tbl.AddRow("Charlie", "35", "London")
    c.Render(tbl)
}
```

Output:
```
┏━━━━━━━━━┳━━━━━┳━━━━━━━━━━━━━━━┓
┃ Name    ┃ Age ┃ City          ┃
┡━━━━━━━━━╇━━━━━╇━━━━━━━━━━━━━━━┩
│ Alice   │ 30  │ New York      │
│ Bob     │ 25  │ San Francisco │
│ Charlie │ 35  │ London        │
└─────────┴─────┴───────────────┘
```

### Column Styling

```go
tbl := table.NewTable()
tbl.AddColumn("Name", table.WithColumnStyle("bold"))
tbl.AddColumn("Age",  table.WithColumnStyle("yellow"), table.WithColumnJustify(console.JustifyCenter))
tbl.AddColumn("City", table.WithColumnStyle("italic"))

tbl.AddRow("[bold]Alice[/]", "30", "[blue]New York[/]")
tbl.AddRow("Bob", "25", "[blue]San Francisco[/]")

c.Render(tbl)
```

### Footer, Sections & Box Styles

```go
tbl := table.NewTableWithOptions(nil,
    table.WithShowFooter(true),
    table.WithBox(box.ROUNDED),
)
tbl.AddColumn("Task",   table.WithColumnStyle("bold"))
tbl.AddColumn("Status", table.WithColumnJustify(console.JustifyCenter))
tbl.AddRow("Setup project", "[green]Done[/]", "Alice")
tbl.AddSection()
tbl.AddRow("Code review",   "[dim]Pending[/]", "Bob")
c.Render(tbl)
```

### Flexible Widths

```go
tbl := table.NewTableWithOptions(nil, table.WithExpand())
tbl.AddColumn("Fixed",    table.WithColumnWidth(12))
tbl.AddColumn("Flex 2x",  table.WithColumnRatio(2))
tbl.AddColumn("Flex 1x",  table.WithColumnRatio(1))
c.Render(tbl)
```

For full table options see the [API Reference](#table) section below.

## Spinners

50+ built-in spinners from [cli-spinners](https://github.com/sindresorhus/cli-spinners):

```go
progress.NewSpinnerColumn(
    progress.WithSpinnerName("dots"),      // dots, line, star, moon, etc.
    progress.WithFinishedText("Done!"),    // Text when complete
)
```

Available spinners: `dots`, `dots2`, `dots3`, `line`, `pipe`, `star`, `hamburger`, `growVertical`, `growHorizontal`, `balloon`, `noise`, `bounce`, `boxBounce`, `triangle`, `arc`, `circle`, `toggle`, `arrow`, `bouncingBar`, `bouncingBall`, `smiley`, `monkey`, `hearts`, `clock`, `earth`, `moon`, `runner`, `pong`, `shark`, and more.

## API Reference

### Progress

```go
// Create and start
p := progress.New(opts...)
p.Start(ctx)
defer p.Stop()

// Task management
taskID := p.AddTask(description, total, opts...)
p.Advance(taskID, amount)
p.Update(taskID, config)
p.RemoveTask(taskID)

// Task timing
p.StartTask(taskID)
p.StopTask(taskID)
p.ResetTask(taskID, start)

// State
p.Finished() bool  // All tasks complete?
p.Refresh()        // Force refresh
```

### TaskUpdateConfig

```go
p.Update(taskID, progress.TaskUpdateConfig{
    Description: &newDesc,
    Total:       &newTotal,
    Completed:   &newCompleted,
    Advance:     &advanceBy,
    Visible:     &isVisible,
    Fields:      map[string]any{"key": "value"},
})
```

### Table

```go
// Create a table
tbl := table.NewTable(headers ...string)
tbl := table.NewTableWithOptions(headers []string, opts ...TableOption)

// Add columns with options
col := tbl.AddColumn("Header",
    table.WithColumnStyle("bold cyan"),
    table.WithColumnJustify(console.JustifyRight),
    table.WithColumnWidth(20),
    table.WithColumnRatio(2),
)

// Add data rows
tbl.AddRow("cell1", "cell2")
tbl.AddStyledRow([]interface{}{"cell1", "cell2"}, style, endSection)
tbl.AddSection()

// Set footer (WithShowFooter must be true)
col.Footer = "Summary"

// Render
c.Render(tbl)
```

### Table Options

```go
table.NewTableWithOptions(headers,
    table.WithTitle("[bold]Title[/]"),
    table.WithCaption("Caption text"),
    table.WithWidth(80),
    table.WithBox(box.ROUNDED),      // Default: box.HEAVY_HEAD
    table.WithExpand(),              // Fill terminal width
    table.WithShowLines(),           // Dividers between all rows
    table.WithShowHeader(false),     // Hide header
    table.WithShowFooter(true),      // Show footer row
    table.WithShowEdge(false),       // Hide outer edges
    table.WithLeading(1),            // Blank lines between rows
    table.WithStyle(s),
    table.WithRowStyles(s1, s2),     // Alternating row styles
    table.WithHeaderStyle(s),
    table.WithFooterStyle(s),
    table.WithBorderStyle(s),
    table.WithPad(0, 1, 0, 1),      // top, right, bottom, left
    table.WithCollapsePadding(),
    table.WithPadEdge(false),
)
```

## Styling

GoRich automatically detects terminal capabilities and uses the best available:
- Truecolor (24-bit) when `COLORTERM=truecolor`
- 256 colors when `TERM` contains `256color`
- Standard 16 colors otherwise
- No color when `NO_COLOR` is set or output is not a terminal

### Progress Bar Colors

The progress bar automatically changes color based on state:
- **In progress**: Magenta
- **Finished**: Green (when task completes)
- **Pulse animation**: Purple gradient (for indeterminate tasks)

## Thread Safety

All `Progress` methods are safe to call from multiple goroutines. Updates are protected by mutexes, and the display refresh happens in a separate goroutine.

## Acknowledgments

- [Rich](https://github.com/Textualize/rich) by Will McGugan - The original Python library
- [cli-spinners](https://github.com/sindresorhus/cli-spinners) by Sindre Sorhus - Spinner definitions
- [go-runewidth](https://github.com/mattn/go-runewidth) - Terminal cell width calculation

## License

MIT License - see LICENSE file
