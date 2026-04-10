# CLAUDE.md - Project Guide for Claude Code

## Project Overview

GoRich is a Go port of Python's [Rich](https://github.com/Textualize/rich) library for beautiful terminal output. It provides:
- **Rich Print** - Styled text with `[bold red]markup[/]` syntax
- **Progress Bars** - Flicker-free, customizable progress displays with multiple concurrent tasks, speed estimation, and various column types

## Architecture

The project follows a layered architecture mirroring Rich's design:

```
gorich/       <- Package-level convenience (Print, Printf, Rule)
        ↓
markup/       <- Rich-style markup parser
  ├── Parse("[bold red]text[/]") -> Text
  └── Render() -> []Segment
        ↓
progress/     <- Progress bar package
  ├── Progress (orchestrator)
  ├── Task (data model with speed estimation)
  ├── Column interface + implementations
  └── ProgressBar (low-level bar renderer)
        ↓
live/         <- Auto-refreshing terminal display
  ├── Live (refresh goroutine, RenderHook)
  └── LiveRender (cursor positioning)
        ↓
console/      <- Terminal output engine
  ├── Console (terminal detection, color system)
  ├── Renderable interface
  └── RenderHook interface
        ↓
segment/      <- Atomic rendering units
  ├── Segment (text + style + control)
  └── Control codes (ANSI cursor movement)
        ↓
style/        <- ANSI styling
  ├── Style (bitmask attributes)
  └── Color (standard/256/truecolor)
```

## Key Design Decisions

### Flicker-Free Rendering
All output is buffered into a single `strings.Builder` and written with one `Write()` call. Never write multiple times per refresh - this causes flicker.

```go
// CORRECT - single write
var buf strings.Builder
buf.WriteString(cursorCodes)
for _, seg := range segments {
    buf.WriteString(seg.Render(colorSys))
}
console.WriteString(buf.String())

// WRONG - multiple writes cause flicker
console.WriteControl(cursorCodes)
for _, seg := range segments {
    console.WriteString(seg.Render(colorSys))
}
```

### Thread Safety & Deadlock Prevention
- `Progress` and `Task` use `sync.Mutex`
- **Critical**: Release mutex before calling into `Live.Start()`/`Live.Stop()` - they call back into `Progress.getRenderable()` which needs the mutex
- Lock ordering: Progress.mu → Task.mu → Live.mu (never reverse)

### Speed Estimation
- Uses a ring buffer of 1000 samples (see `sampleRing` in task.go)
- Window of 30 seconds (configurable via `WithSpeedEstimatePeriod`)
- Speed = (newest.completed - oldest.completed) / (newest.timestamp - oldest.timestamp)

### Style Three-State Attributes
Styles use two bitmasks to support "explicitly off" (not just "not set"):
```go
type Style struct {
    attrs    Attribute  // which attrs are ON
    setAttrs Attribute  // which attrs are explicitly SET
}
// Bit in setAttrs but not attrs = explicitly OFF
// Bit absent from setAttrs = inherit from parent
```

## Package Reference

### progress/
- `Progress` - Main orchestrator, manages tasks and Live display
- `Task` - Individual progress task with timing and speed calculation
- `TaskSnapshot` - Read-only copy for safe rendering (avoids holding locks)
- `Column` interface - All column types implement this
- Column implementations:
  - `DescriptionColumn()` - Task description (right-aligned, supports markup)
  - `BarColumn` - The visual progress bar (switches to green when task finishes)
  - `TaskProgressColumn` - Percentage display
  - `TimeRemainingColumn` - ETA
  - `TimeElapsedColumn` - Elapsed time
  - `SpinnerColumn` - Animated spinner
  - `DownloadColumn` - File size progress (e.g., "5.2 MB/10.0 MB")
  - `TransferSpeedColumn` - Transfer rate
  - `MofNCompleteColumn` - "M/N" display
  - `SeparatorColumn` - Static separator (e.g., "•")
- `Reader`/`Writer` - io.Reader/Writer wrappers for IO progress

### live/
- `Live` - Auto-refreshing display using goroutine + ticker
- `LiveRender` - Tracks rendered shape, generates cursor repositioning codes

### console/
- `Console` - Terminal output with detection (isatty, color system)
- `Renderable` interface - `Render(c *Console, opts Options) []segment.Segment`
- `RenderHook` interface - Intercepts print calls (used by Live)
- `Options` - Rendering constraints (width, color system, etc.)

### segment/
- `Segment` - Atomic unit: `{Text, Style, Control}`
- `Control` - ANSI escape sequences (cursor up, erase line, etc.)
- Helper functions: `SplitLines`, `AdjustLineLength`, `Simplify`, `Divide`

### style/
- `Color` - Supports standard (16), 256-color, and truecolor with downgrading
- `Style` - Text attributes (bold, italic, etc.) with ANSI rendering
- `Parse(string)` - Parse style strings like "bold red on white"

### markup/
- `Parse(string)` - Parse markup string into Text with styled spans
- `Render(string)` - Convenience function: parse and convert to segments
- `Strip(string)` - Remove markup tags, return plain text
- `VisibleLength(string)` - Length of visible text (excluding markup tags)
- `Escape(string)` - Escape text so it won't be interpreted as markup

### spinner/
- `Spinner` - Animated spinner widget
- `Spinners` map - 50+ spinner definitions from cli-spinners

## Common Tasks

### Adding a New Column Type
1. Create struct implementing `Column` interface in `progress/column.go` or `progress/columns_extra.go`
2. Implement `Render(task TaskSnapshot, c *console.Console, opts console.Options) []segment.Segment`
3. Implement `MaxRefresh() time.Duration` (return 0 for no throttling)

### Modifying Default Appearance
- Default columns: `progress/column.go` → `DefaultColumns()`
- Bar characters: `progress/bar.go` → constants at top
- Bar colors: `progress/bar.go` → `getBarStyle()` function (complete=magenta, finished=green, back=gray, pulse=purple)
- Column colors: Search for `style.Parse()` calls in column Render methods

### Testing Progress Display
The refresh happens in a goroutine, so captured output won't show intermediate states. For visual testing, run the examples:
```bash
go run ./example/progress/  # Progress bar demo
go run ./example/print/     # Rich print demo
```

## File Locations

| What | Where |
|------|-------|
| Rich print API | `print.go`, `console/print.go` |
| Markup parser | `markup/markup.go` |
| Main progress API | `progress/progress.go` |
| Task & speed estimation | `progress/task.go` |
| Column implementations | `progress/column.go`, `progress/columns_extra.go` |
| Progress bar renderer | `progress/bar.go` |
| IO wrappers | `progress/reader.go` |
| Live display | `live/live.go`, `live/render.go` |
| Console & terminal | `console/console.go` |
| ANSI control codes | `segment/control.go` |
| Style/color system | `style/style.go`, `style/color.go` |
| Spinner definitions | `spinner/spinners.go` |

## External Dependencies

- `github.com/mattn/go-runewidth` - Cell width calculation for CJK/emoji
- `golang.org/x/term` - Terminal detection and size

## Reference Implementation

The Python Rich source is in `rich/` directory for reference:
- `rich/rich/progress.py` - Progress and Task classes
- `rich/rich/progress_bar.py` - Bar renderer
- `rich/rich/live.py` - Live display
- `rich/rich/console.py` - Console implementation
- `rich/rich/_spinners.py` - Spinner definitions
