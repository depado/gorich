package progress

import (
	"fmt"
	"time"

	"github.com/depado/gorich/console"
	"github.com/depado/gorich/markup"
	"github.com/depado/gorich/segment"
	"github.com/depado/gorich/style"
)

// Column is the interface for progress bar columns.
type Column interface {
	// Render renders the column for the given task.
	Render(task TaskSnapshot, c *console.Console, opts console.Options) []segment.Segment

	// MaxRefresh returns the minimum interval between re-renders, or 0 for no throttle.
	MaxRefresh() time.Duration
}

// TextColumn displays text with optional formatting.
type TextColumn struct {
	Text       string       // Text to display (can use {task.description}, etc.)
	Style      *style.Style
	Justify    console.Justify
	Width      int  // Minimum width (0 = no padding)
	NoWrap     bool
	Markup     bool // Enable markup parsing (default true)
	maxRefresh time.Duration
}

// NewTextColumn creates a new text column.
func NewTextColumn(text string, opts ...func(*TextColumn)) *TextColumn {
	tc := &TextColumn{
		Text:   text,
		Markup: true, // Enable markup by default
	}
	for _, opt := range opts {
		opt(tc)
	}
	return tc
}

// WithTextStyle sets the text style.
func WithTextStyle(s style.Style) func(*TextColumn) {
	return func(tc *TextColumn) {
		tc.Style = &s
	}
}

// WithJustify sets the text justification.
func WithJustify(j console.Justify) func(*TextColumn) {
	return func(tc *TextColumn) {
		tc.Justify = j
	}
}

// WithWidth sets the minimum width for padding.
func WithWidth(w int) func(*TextColumn) {
	return func(tc *TextColumn) {
		tc.Width = w
	}
}

// WithMarkup enables or disables markup parsing (default: true).
func WithMarkup(enabled bool) func(*TextColumn) {
	return func(tc *TextColumn) {
		tc.Markup = enabled
	}
}

// Render implements Column.
func (tc *TextColumn) Render(task TaskSnapshot, c *console.Console, opts console.Options) []segment.Segment {
	text := tc.formatText(task)

	// Calculate visible length (without markup tags) for padding
	visibleLen := len(text)
	if tc.Markup {
		visibleLen = markup.VisibleLength(text)
	}

	// Apply padding based on justification
	if tc.Width > 0 && visibleLen < tc.Width {
		padding := tc.Width - visibleLen
		switch tc.Justify {
		case console.JustifyRight:
			text = fmt.Sprintf("%*s", padding, "") + text
		case console.JustifyCenter:
			left := padding / 2
			right := padding - left
			text = fmt.Sprintf("%*s%s%*s", left, "", text, right, "")
		default: // Left justify (default)
			text = text + fmt.Sprintf("%*s", padding, "")
		}
	}

	// Parse markup if enabled
	if tc.Markup {
		segments := markup.Render(text)
		// Apply base style if set
		if tc.Style != nil {
			for i := range segments {
				if segments[i].Style != nil {
					combined := tc.Style.Add(*segments[i].Style)
					segments[i].Style = &combined
				} else {
					segments[i].Style = tc.Style
				}
			}
		}
		return segments
	}

	return []segment.Segment{segment.NewText(text, tc.Style)}
}

func (tc *TextColumn) formatText(task TaskSnapshot) string {
	// Simple template substitution
	text := tc.Text
	if text == "" {
		text = task.Description
	}
	// In a full implementation, this would support {task.description}, {task.percentage}, etc.
	return text
}

// MaxRefresh implements Column.
func (tc *TextColumn) MaxRefresh() time.Duration {
	return tc.maxRefresh
}

// SeparatorColumn displays a static separator string.
type SeparatorColumn struct {
	Text  string
	Style *style.Style
}

// NewSeparatorColumn creates a separator column.
func NewSeparatorColumn(text string) *SeparatorColumn {
	return &SeparatorColumn{Text: text}
}

// Render implements Column.
func (sc *SeparatorColumn) Render(task TaskSnapshot, c *console.Console, opts console.Options) []segment.Segment {
	return []segment.Segment{segment.NewText(sc.Text, sc.Style)}
}

// MaxRefresh implements Column.
func (sc *SeparatorColumn) MaxRefresh() time.Duration {
	return 0
}

// DescriptionColumn creates a column that displays the task description.
// This is a convenience wrapper around TextColumn with sensible defaults.
// Supports markup syntax like "[bold red]Downloading[/]" in task descriptions.
func DescriptionColumn(opts ...func(*TextColumn)) *TextColumn {
	tc := &TextColumn{
		Text:    "", // Empty means use task.Description
		Justify: console.JustifyRight,
		Width:   12,
		Markup:  true, // Enable markup by default
	}
	for _, opt := range opts {
		opt(tc)
	}
	return tc
}

// BarColumn displays the progress bar.
type BarColumn struct {
	BarWidth      *int   // Width of the bar, nil = auto
	Style         string // Style name for completed portion
	CompleteStyle string
	FinishedStyle string
	PulseStyle    string
	maxRefresh    time.Duration
}

// NewBarColumn creates a new bar column.
func NewBarColumn(opts ...func(*BarColumn)) *BarColumn {
	bc := &BarColumn{}
	for _, opt := range opts {
		opt(bc)
	}
	return bc
}

// WithBarWidth sets the bar width.
func WithBarWidth(width int) func(*BarColumn) {
	return func(bc *BarColumn) {
		bc.BarWidth = &width
	}
}

// Render implements Column.
func (bc *BarColumn) Render(task TaskSnapshot, c *console.Console, opts console.Options) []segment.Segment {
	bar := &ProgressBar{
		Total:         task.Total,
		Completed:     task.Completed,
		Width:         bc.BarWidth,
		Pulse:         !task.Started,
		AnimationTime: task.CurrentTime,
		ASCIIOnly:     opts.ASCIIOnly,
		Finished:      task.Finished,
		BackStyle:     bc.Style,
		CompleteStyle: bc.CompleteStyle,
		FinishedStyle: bc.FinishedStyle,
		PulseStyle:    bc.PulseStyle,
	}
	return bar.Render(c, opts)
}

// MaxRefresh implements Column.
func (bc *BarColumn) MaxRefresh() time.Duration {
	return bc.maxRefresh
}

// TaskProgressColumn displays the progress percentage or speed.
type TaskProgressColumn struct {
	ShowSpeed    bool // Show speed when total is unknown
	maxRefresh   time.Duration
}

// NewTaskProgressColumn creates a new task progress column.
func NewTaskProgressColumn(showSpeed bool) *TaskProgressColumn {
	return &TaskProgressColumn{ShowSpeed: showSpeed}
}

// Render implements Column.
func (tpc *TaskProgressColumn) Render(task TaskSnapshot, c *console.Console, opts console.Options) []segment.Segment {
	var text string

	if task.Total != nil {
		text = fmt.Sprintf("%3.0f%%", task.Percentage)
	} else if tpc.ShowSpeed {
		speed := task.GetSpeed()
		if speed != nil {
			text = formatSpeed(*speed)
		} else {
			text = "---"
		}
	} else {
		text = "---"
	}

	s := style.Parse("magenta")
	return []segment.Segment{segment.NewText(text, &s)}
}

// MaxRefresh implements Column.
func (tpc *TaskProgressColumn) MaxRefresh() time.Duration {
	return tpc.maxRefresh
}

// TimeRemainingColumn displays the estimated time remaining.
type TimeRemainingColumn struct {
	Compact            bool    // Use compact format (MM:SS)
	ElapsedWhenFinished bool   // Show elapsed time when finished
	maxRefresh         time.Duration
}

// NewTimeRemainingColumn creates a new time remaining column.
func NewTimeRemainingColumn() *TimeRemainingColumn {
	return &TimeRemainingColumn{
		maxRefresh: 500 * time.Millisecond, // Throttle to reduce jitter
	}
}

// Render implements Column.
func (trc *TimeRemainingColumn) Render(task TaskSnapshot, c *console.Console, opts console.Options) []segment.Segment {
	var text string

	if task.Finished && trc.ElapsedWhenFinished && task.Elapsed != nil {
		text = formatDuration(*task.Elapsed, trc.Compact)
	} else if task.TimeRemaining != nil {
		text = formatDuration(*task.TimeRemaining, trc.Compact)
	} else {
		text = "-:--:--"
		if trc.Compact {
			text = "--:--"
		}
	}

	s := style.Parse("cyan")
	return []segment.Segment{segment.NewText(text, &s)}
}

// MaxRefresh implements Column.
func (trc *TimeRemainingColumn) MaxRefresh() time.Duration {
	return trc.maxRefresh
}

// TimeElapsedColumn displays the elapsed time.
type TimeElapsedColumn struct {
	Compact    bool
	maxRefresh time.Duration
}

// NewTimeElapsedColumn creates a new time elapsed column.
func NewTimeElapsedColumn() *TimeElapsedColumn {
	return &TimeElapsedColumn{}
}

// Render implements Column.
func (tec *TimeElapsedColumn) Render(task TaskSnapshot, c *console.Console, opts console.Options) []segment.Segment {
	var text string

	if task.Elapsed != nil {
		text = formatDuration(*task.Elapsed, tec.Compact)
	} else {
		text = "-:--:--"
		if tec.Compact {
			text = "--:--"
		}
	}

	s := style.Parse("cyan")
	return []segment.Segment{segment.NewText(text, &s)}
}

// MaxRefresh implements Column.
func (tec *TimeElapsedColumn) MaxRefresh() time.Duration {
	return tec.maxRefresh
}

// MofNCompleteColumn displays "M/N" progress.
type MofNCompleteColumn struct {
	Separator  string
	maxRefresh time.Duration
}

// NewMofNCompleteColumn creates a new M/N column.
func NewMofNCompleteColumn(separator string) *MofNCompleteColumn {
	if separator == "" {
		separator = "/"
	}
	return &MofNCompleteColumn{Separator: separator}
}

// Render implements Column.
func (mc *MofNCompleteColumn) Render(task TaskSnapshot, c *console.Console, opts console.Options) []segment.Segment {
	var text string

	if task.Total != nil {
		// Calculate padding based on total width
		totalStr := fmt.Sprintf("%.0f", *task.Total)
		format := fmt.Sprintf("%%%d.0f%s%%s", len(totalStr), mc.Separator)
		text = fmt.Sprintf(format, task.Completed, totalStr)
	} else {
		text = fmt.Sprintf("%.0f", task.Completed)
	}

	s := style.Parse("green")
	return []segment.Segment{segment.NewText(text, &s)}
}

// MaxRefresh implements Column.
func (mc *MofNCompleteColumn) MaxRefresh() time.Duration {
	return mc.maxRefresh
}

// Helper functions

func formatDuration(seconds float64, compact bool) string {
	if seconds < 0 {
		seconds = 0
	}

	s := int(seconds)
	hours := s / 3600
	minutes := (s % 3600) / 60
	secs := s % 60

	if compact {
		if hours > 0 {
			return fmt.Sprintf("%d:%02d:%02d", hours, minutes, secs)
		}
		return fmt.Sprintf("%02d:%02d", minutes, secs)
	}

	return fmt.Sprintf("%d:%02d:%02d", hours, minutes, secs)
}

func formatSpeed(speed float64) string {
	if speed >= 1e9 {
		return fmt.Sprintf("%.1f G/s", speed/1e9)
	}
	if speed >= 1e6 {
		return fmt.Sprintf("%.1f M/s", speed/1e6)
	}
	if speed >= 1e3 {
		return fmt.Sprintf("%.1f K/s", speed/1e3)
	}
	return fmt.Sprintf("%.1f /s", speed)
}

// DefaultColumns returns the default set of columns.
func DefaultColumns() []Column {
	return []Column{
		DescriptionColumn(),
		NewBarColumn(),
		NewTaskProgressColumn(false),
		NewSeparatorColumn("•"),
		NewTimeRemainingColumn(),
	}
}
