package live

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/depado/gorich/console"
	"github.com/depado/gorich/markup"
	"github.com/depado/gorich/segment"
	"github.com/depado/gorich/spinner"
	"github.com/depado/gorich/style"
)

// BlockStatus tracks the lifecycle of a single block.
type BlockStatus int

const (
	// BlockRunning means the block's command is in progress and the spinner animates.
	BlockRunning BlockStatus = iota
	// BlockSucceeded means the block finished successfully (exit code 0).
	BlockSucceeded
	// BlockFailed means the block's command returned a non-zero exit code.
	BlockFailed
)

// Block is a single entry in a BlockDisplay. It has a header (title + status
// indicator) and a ring buffer of the last maxLines output lines, which stay
// visible after the block finishes.
//
// Blocks are not created directly; use BlockDisplay.Start to add one.
type Block struct {
	Title    string
	Status   BlockStatus
	Lines    []string
	maxLines int
	Elapsed  time.Duration
	ExitCode int

	// SpinnerStyle overrides the running-frame spinner style. nil = dim cyan.
	SpinnerStyle *style.Style
	// RunningStyle is the base style for the running header title. Markup in the
	// title layers on top of it (overriding only what it sets), so use this for a
	// uniform default look and markup for per-title tweaks. nil = unstyled.
	RunningStyle *style.Style
	// SucceededStyle overrides the succeeded header title style. nil = green bold.
	SucceededStyle *style.Style
	// FailedStyle overrides the failed header title style. nil = red bold.
	FailedStyle *style.Style
	// OutputStyle overrides the per-line output style. nil = dim.
	OutputStyle *style.Style

	// spinner is owned by BlockDisplay, not the caller.
	spinner *spinner.Spinner
	start   time.Time
}

// BlockDisplay is a [console.Renderable] that shows a growing list of blocks
// in a stable terminal region via [Live]. Each block has a header line
// (spinner while running, plain title while done) followed by its last N
// output lines. Old blocks scroll off the top once the display height exceeds
// the terminal height.
//
// Use Start to create a new block, BlockWriter to feed it output lines, and
// Finish to set its final status + elapsed time. A BlockDisplay is safe for
// concurrent use: Start/AppendLine/Finish may be called from any goroutine.
type BlockDisplay struct {
	mu           sync.Mutex
	blocks       []*Block
	defaultMax   int
	spinnerName  string
	reserveSpace bool
	prefix       string
}

// BlockDisplayOption configures a BlockDisplay.
type BlockDisplayOption func(*BlockDisplay)

// WithBlockMaxLines sets the default ring-buffer size for new blocks (default 3).
func WithBlockMaxLines(n int) BlockDisplayOption {
	return func(b *BlockDisplay) { b.defaultMax = n }
}

// WithBlockReserveSpace, when enabled, pads each block with blank lines up to
// maxLines so the block height is stable from the moment it starts. Without
// it (the default), blocks grow organically as output arrives and never shrink.
func WithBlockReserveSpace(reserve bool) BlockDisplayOption {
	return func(b *BlockDisplay) { b.reserveSpace = reserve }
}

// WithBlockPrefix sets the string prepended to every output line (default "  "
// — two spaces). Use a vertical bar like "│ " for a tree-style visual border.
func WithBlockPrefix(prefix string) BlockDisplayOption {
	return func(b *BlockDisplay) { b.prefix = prefix }
}

// WithBlockSpinnerName sets the spinner animation used for running blocks (default "dots").
func WithBlockSpinnerName(name string) BlockDisplayOption {
	return func(b *BlockDisplay) { b.spinnerName = name }
}

// NewBlockDisplay creates a BlockDisplay.
func NewBlockDisplay(opts ...BlockDisplayOption) *BlockDisplay {
	b := &BlockDisplay{
		defaultMax:  3,
		spinnerName: "dots",
		prefix:      "  ",
	}
	for _, opt := range opts {
		opt(b)
	}
	return b
}

// Start appends a new running block for title and returns its index. Callers
// must not mutate the returned *Block's Lines/Status directly; use AppendLine
// and Finish instead.
func (d *BlockDisplay) Start(title string) int {
	blk := &Block{
		Title:    title,
		Status:   BlockRunning,
		Lines:    make([]string, 0, d.defaultMax),
		maxLines: d.defaultMax,
		spinner:  spinner.New(d.spinnerName),
		start:    time.Now(),
	}
	d.mu.Lock()
	d.blocks = append(d.blocks, blk)
	d.mu.Unlock()
	return len(d.blocks) - 1
}

// AppendLine adds an output line to block idx, trimming the ring buffer to
// maxLines. Lines are dimmed when rendered.
func (d *BlockDisplay) AppendLine(idx int, line string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if idx < 0 || idx >= len(d.blocks) {
		return
	}
	blk := d.blocks[idx]
	blk.Lines = append(blk.Lines, line)
	if len(blk.Lines) > blk.maxLines {
		blk.Lines = blk.Lines[len(blk.Lines)-blk.maxLines:]
	}
}

// Finish marks the block as Succeeded (code 0) or Failed (non-zero) and
// records its elapsed time. Safe to call multiple times; the first call wins.
func (d *BlockDisplay) Finish(idx int, exitCode int) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if idx < 0 || idx >= len(d.blocks) {
		return
	}
	blk := d.blocks[idx]
	if blk.Status != BlockRunning {
		return
	}
	blk.Elapsed = time.Since(blk.start)
	blk.ExitCode = exitCode
	if exitCode == 0 {
		blk.Status = BlockSucceeded
	} else {
		blk.Status = BlockFailed
	}
}

// AppendLines is a convenience helper wrapping AppendLine for a multi-line
// string without a trailing newline.
func (d *BlockDisplay) AppendLines(idx int, s string) {
	s = strings.TrimRight(s, "\n")
	if s == "" {
		return
	}
	for line := range strings.SplitSeq(s, "\n") {
		d.AppendLine(idx, line)
	}
}

// BlockWriter is an [io.Writer] that buffers partial lines and flushes
// complete lines to [BlockDisplay.AppendLine]. Create one per block with
// [BlockDisplay.NewWriter] and use it as the stdout/stderr sink for a child
// process. It is safe for concurrent use within a single block.
type BlockWriter struct {
	display *BlockDisplay
	idx     int
	mu      sync.Mutex
	buf     []byte
}

// NewWriter returns an io.Writer that appends completed lines to block idx.
func (d *BlockDisplay) NewWriter(idx int) *BlockWriter {
	return &BlockWriter{display: d, idx: idx}
}

// Write implements io.Writer. It is safe for concurrent use.
func (w *BlockWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.buf = append(w.buf, p...)
	for {
		i := bytes.IndexByte(w.buf, '\n')
		if i < 0 {
			break
		}
		line := string(w.buf[:i])
		w.buf = w.buf[i+1:]
		line = strings.TrimRight(line, "\r")
		w.display.AppendLine(w.idx, line)
	}
	return len(p), nil
}

// Render implements console.Renderable.
func (d *BlockDisplay) Render(c *console.Console, opts console.Options) []segment.Segment {
	now := float64(time.Now().UnixNano()) / 1e9

	d.mu.Lock()
	defer d.mu.Unlock()

	var lines [][]segment.Segment
	for _, blk := range d.blocks {
		// Header
		lines = append(lines, renderBlockHeader(blk, now))

		// Output lines (kept after finish, not collapsed). The ring buffer
		// caps at maxLines so the block grows as output arrives but never
		// exceeds maxLines.
		// Output lines (kept after finish, not collapsed). The ring buffer
		// caps at maxLines so the block grows as output arrives but never
		// exceeds maxLines.
		prefixSegs := markup.Render(d.prefix)
		fallback := defaultOutputStyle(blk)
		for i := range prefixSegs {
			if prefixSegs[i].Style == nil {
				prefixSegs[i].Style = fallback
			}
		}
		for _, line := range blk.Lines {
			lineSegs := markup.Render(line)
			for i := range lineSegs {
				if lineSegs[i].Style == nil {
					lineSegs[i].Style = fallback
				}
			}
			row := make([]segment.Segment, 0, len(prefixSegs)+len(lineSegs))
			row = append(row, prefixSegs...)
			row = append(row, lineSegs...)
			lines = append(lines, row)
		}
		// When reserveSpace is enabled, pad with blank lines up to maxLines
		// so the block height is stable from the start.
		if d.reserveSpace {
			for j := len(blk.Lines); j < blk.maxLines; j++ {
				lines = append(lines, []segment.Segment{})
			}
		}
	}

	// Crop to terminal height: keep the bottom N lines so old blocks scroll off.
	maxHeight := opts.Size.Height
	if maxHeight > 0 && len(lines) > maxHeight {
		lines = lines[len(lines)-maxHeight:]
	}

	// Truncate each line to the terminal width so a long line never wraps onto
	// a second physical row. The Live cursor math counts one row per logical
	// line, so a wrapped line would leave stale rows the erase can't reach,
	// stacking duplicate frames on every refresh.
	// ponytail: measures with runewidth (tab = 0 cells); a tab-heavy line could
	// still wrap. Expand tabs before truncating if that surfaces.
	width := opts.Size.Width

	var segs []segment.Segment
	for i, line := range lines {
		if i > 0 {
			segs = append(segs, segment.Segment{Text: "\n"})
		}
		if width > 0 {
			line = segment.AdjustLineLength(line, width, false)
		}
		segs = append(segs, line...)
	}
	return segs
}

func renderBlockHeader(blk *Block, now float64) []segment.Segment {
	switch blk.Status {
	case BlockSucceeded:
		s := defaultSucceededStyle(blk)
		line := []segment.Segment{segment.NewText("✓ ", s)}
		line = append(line, renderTitle(blk.Title, s)...)
		if blk.Elapsed > 0 {
			line = append(line, segment.NewText(fmt.Sprintf(" (%s)", blk.Elapsed.Round(time.Millisecond)), &style.Dim))
		}
		return line
	case BlockFailed:
		s := defaultFailedStyle(blk)
		line := []segment.Segment{segment.NewText("✗ ", s)}
		line = append(line, renderTitle(blk.Title, s)...)
		line = append(line, segment.NewText(fmt.Sprintf(" (exit %d)", blk.ExitCode), s))
		if blk.Elapsed > 0 {
			line = append(line, segment.NewText(fmt.Sprintf(" (%s)", blk.Elapsed.Round(time.Millisecond)), &style.Dim))
		}
		return line
	default: // BlockRunning
		spinSegs := blk.spinner.Render(now)
		styleOverride := defaultSpinnerStyle(blk)
		if styleOverride != nil {
			for i := range spinSegs {
				if spinSegs[i].Style == nil {
					spinSegs[i].Style = styleOverride
				}
			}
		}
		titleStyle := defaultRunningStyle(blk)
		line := append(spinSegs, segment.NewText(" ", titleStyle))
		return append(line, renderTitle(blk.Title, titleStyle)...)
	}
}

// renderTitle parses the title as markup and composes each span with the base
// style: fallback is the base layer, markup layers on top via Style.Add so it
// overrides only the attributes it sets and inherits the rest. A plain title
// (no markup) is rendered entirely in fallback.
func renderTitle(title string, fallback *style.Style) []segment.Segment {
	segs := markup.Render(title)
	for i := range segs {
		if segs[i].Style == nil {
			segs[i].Style = fallback
		} else if fallback != nil {
			merged := fallback.Add(*segs[i].Style)
			segs[i].Style = &merged
		}
	}
	return segs
}

func defaultSpinnerStyle(b *Block) *style.Style {
	if b.SpinnerStyle != nil {
		return b.SpinnerStyle
	}
	s := style.New().WithForeground(style.StandardColor(6)) // cyan
	return &s
}

func defaultRunningStyle(b *Block) *style.Style {
	return b.RunningStyle
}

func defaultSucceededStyle(b *Block) *style.Style {
	if b.SucceededStyle != nil {
		return b.SucceededStyle
	}
	s := style.New().WithForeground(style.StandardColor(2)).WithBold(true) // green bold
	return &s
}

func defaultFailedStyle(b *Block) *style.Style {
	if b.FailedStyle != nil {
		return b.FailedStyle
	}
	s := style.New().WithForeground(style.StandardColor(1)).WithBold(true) // red bold
	return &s
}

func defaultOutputStyle(b *Block) *style.Style {
	if b.OutputStyle != nil {
		return b.OutputStyle
	}
	return &style.Dim
}

var _ io.Writer = (*BlockWriter)(nil)
