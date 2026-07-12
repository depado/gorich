# GoRich

A Go port of Python's [Rich](https://github.com/Textualize/rich) library for beautiful terminal output. Features styled text with markup syntax, progress bars, tables, and customizable displays.

## Features

- **Rich Print** - Styled text with `[bold red]markup[/]` syntax
- **Progress Bars** - Multiple concurrent tasks with customizable columns and per-group sections
- **Live Blocks** - Growing per-task output blocks with animated spinners, styled line prefixes, and optional space reservation
- **Tables** - Bordered tables with column styling, markup cells, row styles, footers, sections, and 19 box styles
- **Flicker-free** - Single-write buffered output with synchronized-output (DEC 2026) frames for smooth updates
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
Processing  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━  65% • 0:00:03
Uploading   ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━  80% • 0:00:01
```

Descriptions are left-aligned and automatically sized to the widest one so the
bars line up. Pass `progress.WithJustify(console.JustifyRight)` for the classic
right-aligned look, or `progress.WithWidth(n)` for a fixed width:

```go
progress.DescriptionColumn(progress.WithJustify(console.JustifyRight))
progress.DescriptionColumn(progress.WithWidth(20))
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
| `DescriptionColumn()` | Task description, markup + auto-width (left-aligned) | `Downloading` |
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

## Marking Tasks as Done

Use `Done()` to explicitly mark a task as finished and optionally update its description:

```go
task := p.AddTask("Waiting for API...", nil)  // indeterminate task
p.Done(task, "[green]Connected![/]")          // mark done + update description
```

`Done()` works on both indeterminate and determinate tasks. For determinate tasks
that haven't reached 100%, it also sets completed to total for a proper visual
finish (green bar, checkmark in spinner, etc.).

## Sections

By default all tasks share one column layout. Use **sections** to give groups of
tasks their own columns and indentation - for example a full summary bar on top
with lightweight worker rows underneath:

```go
p := progress.New(
    progress.WithColumns(                     // default section (section 0)
        progress.NewSpinnerColumn(),
        progress.DescriptionColumn(),
        progress.NewBarColumn(),
        progress.NewTaskProgressColumn(false),
        progress.NewSeparatorColumn("•"),
        progress.NewTimeRemainingColumn(),
    ),
)

// A second section with fewer columns, indented by 2 spaces
workers := p.AddSection(
    progress.WithSectionColumns(
        progress.NewSpinnerColumn(),
        progress.DescriptionColumn(),
    ),
    progress.WithSectionIndent(2),
)

p.Start(context.Background())
defer p.Stop()

total := 88.0
summary := p.AddTask("[bold]Syncing depado[/]", &total)   // goes to section 0
w1 := workers.AddTask("[cyan]articles - cloning...[/]", nil)  // goes to workers

// Every task is addressed by its TaskID regardless of section
p.Advance(summary, 1)
p.Done(w1, "[green]articles[/]")
```

Output:
```
⠸ Syncing depado ━━━━━━━━━━━━━━━━━━━━  45% • 0:01:23
  ⠸ articles - cloning...
  ⠸ buoy - pulling...
  ✓ gorich
```

Sections are a rendering grouping only: task updates always go through the
`Progress` methods by `TaskID`. If you never call `AddSection`, behavior is
identical to a single default section. Each section auto-sizes its own
`DescriptionColumn` independently.

## Live Blocks

`BlockDisplay` is a `console.Renderable` for showing a growing list of
output blocks — useful when running parallel commands and streaming their
output in a live, in-place display. Each block has a header with an animated
spinner while running (which freezes into ✓ or ✗ on completion), followed by
its last N output lines. Wrap it in a `live.Live` for auto-refresh:

```go
package main

import (
    "context"
    "os/exec"
    "sync"
    "time"

    "github.com/depado/gorich/console"
    "github.com/depado/gorich/live"
)

func main() {
    c := console.New()
    display := live.NewBlockDisplay(
        live.WithBlockMaxLines(3),
        live.WithBlockSpinnerName("dots"),
        live.WithBlockPrefix("[blue]│ [/]"),
        live.WithBlockReserveSpace(false),
    )

    l := live.New(c, display, live.WithAutoRefresh(true), live.WithRefreshRate(15))
    l.Start(context.Background())
    defer l.Stop()

    var wg sync.WaitGroup
    repos := []struct{ name, expr string }{
        {"alpha", "echo hi && sleep 0.5 && echo there && sleep 0.3 && echo done"},
        {"delta", "echo failing && echo badly && sleep 0.3 && exit 2"},
    }
    for _, r := range repos {
        wg.Add(1)
        go func(r struct{ name, expr string }) {
            defer wg.Done()
            idx := display.Start(r.name)
            out := display.NewWriter(idx)
            cmd := exec.Command("sh", "-c", r.expr)
            cmd.Stdout = out
            cmd.Stderr = out
            if err := cmd.Run(); err != nil {
                if e, ok := err.(*exec.ExitError); ok {
                    display.Finish(idx, e.ExitCode())
                } else {
                    display.Finish(idx, 1)
                }
            } else {
                display.Finish(idx, 0)
            }
        }(r)
    }
    wg.Wait()
    time.Sleep(300 * time.Millisecond)
}
```

Output (in a terminal):

```
⠹ alpha
│ hi
│ there
✗ delta (exit 2) (312ms)
│ failing
│ badly
✓ alpha (807ms)
│ hi
│ there
│ done
```

### Block options

| Option | Default | Description |
|--------|---------|-------------|
| `WithBlockMaxLines(n)` | `3` | Max output lines kept per block (ring buffer) |
| `WithBlockSpinnerName(name)` | `"dots"` | Spinner animation for running blocks (see [Spinners](#spinners)) |
| `WithBlockPrefix(prefix)` | `"  "` | String prepended to every output line. Supports [markup](#markup-syntax) — e.g. `"[blue]│ [/]"` renders a blue vertical bar |
| — | — | Output lines passed to `AppendLine` / `BlockWriter` also support markup. Use `"[red]error[/]"` for stderr, `"[dim]output[/]"` for stdout, or plain text (inherits the block's `OutputStyle`) |
| — | — | Block titles passed to `Start` also support markup — e.g. `display.Start("[cyan]api[/]")`. Markup layers over the status style, so `[white]auth[/]` on a finished block keeps the status bold and only overrides the color |
| `WithBlockReserveSpace(bool)` | `false` | When true, pads each block with blank lines up to `maxLines` so the height is stable from the start. When false (default), blocks grow organically as output arrives |

### Per-block style overrides

Each `Block` (obtained indirectly via `Start`) has optional style fields that
override the defaults:

```go
type Block struct {
    SpinnerStyle   *style.Style  // nil = cyan
    RunningStyle   *style.Style  // nil = unstyled; base style, markup in the title layers on top
    SucceededStyle *style.Style  // nil = green bold
    FailedStyle    *style.Style  // nil = red bold
    OutputStyle    *style.Style  // nil = dim
    // ...plus Title, Status, Lines, Elapsed, ExitCode
}
```

### API

```go
display := live.NewBlockDisplay(opts...)

// Lifecycle
idx := display.Start("repo-name")        // append a running block, get its index
display.AppendLine(idx, "output line")    // add a line (ring-buffered to maxLines)
display.Finish(idx, exitCode)             // mark done (0=succeeded, non-zero=failed)

// IO
writer := display.NewWriter(idx)          // io.Writer that flushes complete lines
```

`BlockDisplay` implements `console.Renderable` and is safe for concurrent use
— `Start`, `AppendLine`, `Finish`, and `BlockWriter.Write` may be called from
any goroutine.

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
p.Done(taskID, description...)  // Mark task as finished
p.RemoveTask(taskID)

// Sections (per-group column layouts)
section := p.AddSection(
    progress.WithSectionColumns(cols...),
    progress.WithSectionIndent(2),
)
taskID := section.AddTask(description, total, opts...)

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

### Predefined Styles

Pre-defined style variables avoid the need to parse strings:

```go
// Attribute styles (use with `&style.Bold`, etc.)
style.Bold      style.Dim       style.Italic
style.Underline style.Blink     style.Reverse
style.Strike    style.Conceal   style.Overline

// Color styles
style.Red     style.Green    style.Blue
style.Yellow  style.Cyan     style.Magenta
style.White   style.Black
```

Used with table/progress APIs that take `*style.Style`:

```go
table.WithHeaderStyle(&style.Bold)
table.WithBorderStyle(&style.Cyan)
```

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
