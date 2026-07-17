package live

import (
	"time"

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

	ejected bool // removed from live display

	// Internal: set by Render when a finished block is displayed as
	// collapsed, cleared by PopEjects after rendering the scrollback
	// commit. Ensures finished blocks are visible at least once before
	// being ejected.
	collapseDisplayed bool
}

// BlockOption configures a block at creation time.
type BlockOption func(*Block)

// WithBlockSpinnerStyle overrides the spinner style for this block.
func WithBlockSpinnerStyle(s style.Style) BlockOption {
	return func(b *Block) { b.SpinnerStyle = &s }
}

// WithBlockRunningStyle sets the base style for the running title.
func WithBlockRunningStyle(s style.Style) BlockOption {
	return func(b *Block) { b.RunningStyle = &s }
}

// WithBlockSucceededStyle overrides the succeeded header style.
func WithBlockSucceededStyle(s style.Style) BlockOption {
	return func(b *Block) { b.SucceededStyle = &s }
}

// WithBlockFailedStyle overrides the failed header style.
func WithBlockFailedStyle(s style.Style) BlockOption {
	return func(b *Block) { b.FailedStyle = &s }
}

// WithBlockOutputStyle overrides the output line style.
func WithBlockOutputStyle(s style.Style) BlockOption {
	return func(b *Block) { b.OutputStyle = &s }
}
