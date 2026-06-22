# CLAUDE.md - Project Guide for Claude Code

## Project Overview

GoRich is a Go port of Python's [Rich](https://github.com/Textualize/rich) library for beautiful terminal output. It provides:
- **Rich Print** - Styled text with `[bold red]markup[/]` syntax
- **Progress Bars** - Flicker-free, customizable progress displays with multiple concurrent tasks, speed estimation, and various column types
- **Tables** - Bordered tables with column styling, alignment, markup cells, row styles, footers, sections, and 16 box styles

## Architecture

The project follows a layered architecture mirroring Rich's design:

```
gorich/       <- Package-level convenience (Print, Printf, Rule, NewTable)
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
table/        <- Table rendering package
  ├── Table (orchestrator, Renderable)
  ├── Column (header, footer, style, justify, vertical, ratio)
  ├── Row (style, end_section)
  ├── render.go (rendering pipeline, column width calculation)
  ├── ratio.go (ratio_resolve, ratio_reduce, ratio_distribute)
  ├── box/ (18 predefined border styles)
  ├── padding/ (cell padding renderable wrapper)
  └── align/ (horizontal + vertical alignment renderable)
        ↓
live/         <- Auto-refreshing terminal display
  ├── Live (refresh goroutine, RenderHook)
  └── LiveRender (cursor positioning)
        ↓
console/      <- Terminal output engine
  ├── Console (terminal detection, color system)
  ├── Renderable interface
  ├── RenderLines() (render to lines for table cell layout)
  └── RenderHook interface
        ↓
segment/      <- Atomic rendering units
  ├── Segment (text + style + control)
  ├── SplitLines, AdjustLineLength, ApplyStyle, Simplify, Divide
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
- `Progress`, `Task`, `SpinnerColumn`, and `Spinner` use `sync.Mutex`
- `Reader.read` and `Writer.written` use `sync/atomic` for concurrent access
- **Critical**: `console.Render()` releases its mutex before calling hooks - hooks may call back into Console methods
- **Critical**: Release Progress mutex before calling into `Live.Start()`/`Live.Stop()` - they call back into `Progress.getRenderable()` which needs the mutex
- Lock ordering: Console.mu must not be held when calling RenderHook methods

### Speed Estimation
- Uses a ring buffer of 1000 samples (see `sampleRing` in task.go)
- Window of 30 seconds (configurable via `WithSpeedEstimatePeriod`)
- Speed = (newest.completed - oldestInWindow.completed) / (newest.timestamp - oldestInWindow.timestamp)
- Only samples within the speed window are used for calculation

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

### Table Rendering Pipeline
Tables render in two phases: measurement (width calculation) and rendering.

**Measurement phase** (`Measure()` / `calculateColumnWidths()`):
1. Wrap each cell in `Padding` via `_getCells()`
2. Measure each column's padded cells → `Measurement(min, max)`
3. If `expand` + ratios: distribute flexible width via `ratioDistribute()`
4. If too wide: collapse widest wrappable columns via `_collapseWidths()` + `ratioReduce()`
5. If too narrow and `expand`: expand all proportionally
6. Column widths include: content width + padding width (`paddingWidth()`)

**Rendering phase** (`_render()`):
1. Substitute box for platform safety (ASCII-only, headed→plain substitutions)
2. Emit top border line via `box.GetTop(widths)`
3. For each row of cells (header, data, footer):
   - Render every cell via `console.RenderLines()` into `[][]Segment`
   - Truncate each line to column width via `segment.AdjustLineLength()`
   - Find max height across cells in the row
   - Vertical alignment: header=bottom, footer=top, data=column.vertical
   - Pad cells to uniform width/height via `setShape()` + `alignVert()`
   - Emit: left_border + cells (with vertical dividers) + right_border + newline per line
   - Emit dividers: after header, between sections, when `showLines`, or `leading > 0`
4. Emit bottom border (before footer row if footer exists, added to match Python Rich)
5. Cell styles: `cell.style` (from `_getCells`) + `row_style` (alternating + per-row)
6. Divider segments inherit row background style for visual continuity

### Padding Collapse Rules
When `collapsePadding=true`:
- Non-first columns: `left = max(0, left - right)` 
- Non-last rows: `bottom = max(0, top - bottom)`

When `padEdge=false`:
- First column: left=0, Last column: right=0
- First row: top=0, Last row: bottom=0

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
- `ErrNotSeekable` - Returned by `Reader.Seek` when underlying reader doesn't support seeking

### table/
- `Table` - Main orchestrator, implements `console.Renderable` + `console.Measurable`
  - Constructor: `NewTable(headers ...string)` or `NewTableWithOptions(headers []string, opts ...TableOption)`
  - Methods: `AddColumn(header, opts)`, `AddRow(values...)`, `AddStyledRow(values, style, endSection)`, `AddSection()`
  - Table options: `WithTitle`, `WithCaption`, `WithWidth`, `WithMinWidth`, `WithBox`, `WithExpand`, `WithShowLines`, `WithShowHeader`, `WithShowFooter`, `WithShowEdge`, `WithLeading`, `WithStyle`, `WithRowStyles`, `WithHeaderStyle`, `WithFooterStyle`, `WithBorderStyle`, `WithPad`, `WithCollapsePadding`, `WithPadEdge`
- `Column` - Column definition (public fields for direct mutation)
  - Fields: `Header`, `Footer`, `HeaderStyle`, `FooterStyle`, `Style`, `Justify`, `Vertical`, `Overflow`, `Width`, `MinWidth`, `MaxWidth`, `Ratio`, `NoWrap`
  - Column options: `WithColumnStyle(s)`, `WithColumnHeaderStyle(s)`, `WithColumnFooterStyle(s)`, `WithColumnJustify(j)`, `WithColumnVertical(v)`, `WithColumnWidth(w)`, `WithColumnMinWidth(w)`, `WithColumnMaxWidth(w)`, `WithColumnRatio(r)`, `WithColumnNoWrap()`, `WithColumnOverflow(o)`
- `Row` - Per-row metadata: `Style`, `EndSection`
- `cellRender` (internal) - Pairs a renderable with display style + vertical alignment
- `markupRenderable` (internal) - Wraps a string, renders through markup parser, implements `console.Measurable`

**Column width calculation** (`render.go`):
- `_measureColumn()` - Measures a column's min/max width from all padded cells
- `_calculateColumnWidths()` - The main algorithm: measure, expand by ratios, collapse if too wide, expand if too narrow
- `_collapseWidths()` - Iteratively reduces widest wrappable columns to fit max_width
- `_getCells()` - Collects all cells for a column: `[header] + row_cells + [footer]`, wrapped in `Padding`
- `paddingWidth()` - Extra column width from padding (accounts for padEdge/collapsePadding)
- Fixed-width columns: `column.Width > 0` → measurement = `(width+padding, width+padding)` — never grow/shrink
- `column.NoWrap` columns: excluded from collapse candidates (`wrapable` list)
- `column.Ratio > 0`: column is flexible, participates in ratio distribution

**Rendering pipeline** (`render.go`):
- `_render()` - Main render: top border → row loop (render cells, align, emit) → bottom border
- `alignVert()` - Vertical alignment helper (top/middle/bottom) using blank lines
- `setShape()` - Pads cell lines to uniform width

### table/box/
- `Box` - Border drawing characters defined by 8 lines × 4 runes (top, head, head_row, mid, row, foot_row, foot, bottom)
- 18 predefined boxes: `ASCII`, `ASCII2`, `ASCII_DOUBLE_HEAD`, `SQUARE`, `SQUARE_DOUBLE_HEAD`, `MINIMAL`, `MINIMAL_HEAVY_HEAD`, `MINIMAL_DOUBLE_HEAD`, `SIMPLE`, `SIMPLE_HEAD`, `SIMPLE_HEAVY`, `HORIZONTALS`, `ROUNDED`, `HEAVY`, `HEAVY_EDGE`, `HEAVY_HEAD`, `DOUBLE`, `DOUBLE_EDGE`, `MARKDOWN`
- Default: `HEAVY_HEAD`
- Key methods: `GetTop(widths)`, `GetRow(widths, level)`, `GetBottom(widths)`, `Substitute(opts)`, `GetPlainHeadedBox()`, `NoEdge()`
- Levels for `GetRow`: `"head"`, `"row"`, `"mid"`, `"foot"`
- Platform safety: `Substitute()` checks `Options.ASCIIOnly` → returns ASCII
- Plain-headed substitutions (when `showHeader=false`): `HEAVY_HEAD→SQUARE`, `ASCII_DOUBLE_HEAD→ASCII2`, etc.

### table/padding/
- `Padding` - Renders content wrapped with CSS-style padding (1/2/4 ints)
- `NewPadding(renderable, pad, style, expand)` - Constructor
- `NewIndent(renderable, level)` - Quick left-indentation
- Implements `console.Renderable` + `console.Measurable`

### table/align/
- `Align` - Horizontal + vertical alignment within a given width/height
- `Method` constants: `Left`, `Center`, `Right`
- `VerticalMethod` constants: `Top`, `Middle`, `Bottom`
- `New(renderable, align, vertical, style, pad, width, height)`

### live/
- `Live` - Auto-refreshing display using goroutine + ticker
- `LiveRender` - Tracks rendered shape, generates cursor repositioning codes

### console/
- `Console` - Terminal output with detection (isatty, color system)
- `Renderable` interface - `Render(c *Console, opts Options) []segment.Segment`
- `RenderLines(r, opts) [][]segment.Segment` - Renders into lines for table cell layout
- `RenderHook` interface - Intercepts render calls (used by Live)
- `Options` - Rendering constraints (width, color system, etc.)
- Console options:
  - `WithWriter(w)` - Set output writer
  - `WithColorSystem(cs)` - Set color system
  - `WithForceTerminal(bool)` - Force terminal mode
  - `WithNoColor(bool)` - Disable colors
  - `WithWidth(int)` - Set fixed width
  - `WithEnviron(map[string]string)` - Custom environment for testing (like Python Rich's `_environ`)

### segment/
- `Segment` - Atomic unit: `{Text, Style, Control}`
- `Control` - ANSI escape sequences (cursor up, erase line, etc.)
- Helper functions: `SplitLines`, `AdjustLineLength`, `Simplify`, `Divide`, `ApplyStyle`, `TotalCellLength`

### style/
- `Color` - Supports standard (16), 256-color, and truecolor with downgrading
- `Style` - Text attributes (bold, italic, etc.) with ANSI rendering
- `Parse(string)` - Parse style strings like "bold red on white"

### markup/
- `Parse(string)` - Parse markup string into Text with styled spans
- `Render(string)` - Convenience function: parse and convert to segments
- `Strip(string)` - Remove markup tags, return plain text
- `VisibleLength(string)` - Terminal cell width of visible text (handles CJK/emoji)
- `Escape(string)` - Escape text so it won't be interpreted as markup

### spinner/
- `Spinner` - Animated spinner widget
- `Spinners` map - 50+ spinner definitions from cli-spinners

## Common Tasks

### Adding a New Column Type
1. Create struct implementing `Column` interface in `progress/column.go` or `progress/columns_extra.go`
2. Implement `Render(task TaskSnapshot, c *console.Console, opts console.Options) []segment.Segment`
3. Implement `MaxRefresh() time.Duration` (return 0 for no throttling)

### Adding a New Box Style
1. Define an 8-line string with 4 characters per line in `table/box/box.go`
2. Follow the pattern: top, head, head_row, mid, row, foot_row, foot, bottom
3. Add to `legacyWindowsSubstitutions` / `plainHeadedSubstitutions` maps if applicable
4. Add to the example and test in `TestAllPredefinedBoxes`

### Modifying Default Appearance
- Default columns: `progress/column.go` → `DefaultColumns()`
- Bar characters: `progress/bar.go` → constants at top
- Bar colors: `progress/bar.go` → `getBarStyle()` function (complete=magenta, finished=green, back=gray, pulse=purple)
- Column colors: Search for `style.Parse()` calls in column Render methods
- Table default box: `table/table.go` → `NewTable()` (defaults to `box.HEAVY_HEAD`)
- Table default padding: `table/table.go` → `NewTable()` (defaults to `padRight: 1, padLeft: 1`)

### Testing Progress Display
The refresh happens in a goroutine, so captured output won't show intermediate states. For visual testing, run the examples:
```bash
go run ./example/progress/  # Progress bar demo
go run ./example/print/     # Rich print demo
go run ./example/table/     # Table demo
```

### Testing Table Rendering
Use `console.WithWidth(n)` and `console.WithNoColor(true)` for deterministic test output:
```go
c := console.New(console.WithWidth(80), console.WithNoColor(true), console.WithForceTerminal(true))
tbl := table.NewTable("Col1", "Col2")
tbl.AddRow("val1", "val2")
result := tbl.Render(c, c.Options())
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
| Table core | `table/table.go`, `table/column.go` |
| Table rendering pipeline | `table/render.go` |
| Table ratio math | `table/ratio.go` |
| Box border styles | `table/box/box.go` |
| Cell padding wrapper | `table/padding/padding.go` |
| Alignment renderable | `table/align/align.go` |
| Live display | `live/live.go`, `live/render.go` |
| Console & terminal | `console/console.go` |
| RenderLines helper | `console/console.go` (line 331) |
| ANSI control codes | `segment/control.go` |
| Segment primitives | `segment/segment.go` |
| Style/color system | `style/style.go`, `style/color.go` |
| Spinner definitions | `spinner/spinners.go` |
| Cell width utilities | `internal/cells/cells.go` |

## External Dependencies

- `github.com/mattn/go-runewidth` - Cell width calculation for CJK/emoji
- `golang.org/x/term` - Terminal detection and size

## Reference Implementation

The Python Rich source is in `rich/` directory for reference:
- `rich/rich/progress.py` - Progress and Task classes
- `rich/rich/progress_bar.py` - Bar renderer
- `rich/rich/live.py` - Live display
- `rich/rich/console.py` - Console implementation
- `rich/rich/table.py` - Table, Column, Row, _Cell classes
- `rich/rich/box.py` - Box drawing characters + 18 styles
- `rich/rich/padding.py` - Cell padding renderable
- `rich/rich/align.py` - Alignment renderable
- `rich/rich/_ratio.py` - Ratio distribution math
- `rich/rich/_loop.py` - Iteration helpers (loop_first_last, loop_last)
- `rich/rich/measure.py` - Measurement(min, max) named tuple
- `rich/rich/segment.py` - Segment, align_top/middle/bottom, set_shape
- `rich/rich/cells.py` - Unicode cell width utilities
