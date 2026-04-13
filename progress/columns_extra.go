package progress

import (
	"fmt"
	"sync"
	"time"

	"github.com/depado/gorich/console"
	"github.com/depado/gorich/segment"
	"github.com/depado/gorich/spinner"
	"github.com/depado/gorich/style"
)

// SpinnerColumn displays an animated spinner.
type SpinnerColumn struct {
	SpinnerName  string
	Style        *style.Style
	FinishedText string
	Speed        float64
	mu           sync.Mutex // protects spinners and perTaskNames maps
	spinners     map[TaskID]*spinner.Spinner
	perTaskNames map[TaskID]string
	maxRefresh   time.Duration
}

// NewSpinnerColumn creates a new spinner column.
func NewSpinnerColumn(opts ...func(*SpinnerColumn)) *SpinnerColumn {
	sc := &SpinnerColumn{
		SpinnerName:  "dots",
		FinishedText: "✓",
		Speed:        1.0,
		spinners:     make(map[TaskID]*spinner.Spinner),
		perTaskNames: make(map[TaskID]string),
	}
	for _, opt := range opts {
		opt(sc)
	}
	return sc
}

// SetTaskSpinner assigns a specific spinner name to a task, overriding the
// column default. Must be called before the task's first render.
func (sc *SpinnerColumn) SetTaskSpinner(id TaskID, name string) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.perTaskNames[id] = name
}

// WithSpinnerName sets the spinner name.
func WithSpinnerName(name string) func(*SpinnerColumn) {
	return func(sc *SpinnerColumn) {
		sc.SpinnerName = name
	}
}

// WithSpinnerStyle sets the spinner style.
func WithSpinnerStyle(s style.Style) func(*SpinnerColumn) {
	return func(sc *SpinnerColumn) {
		sc.Style = &s
	}
}

// WithFinishedText sets the text shown when task is finished.
func WithFinishedText(text string) func(*SpinnerColumn) {
	return func(sc *SpinnerColumn) {
		sc.FinishedText = text
	}
}

// Render implements Column.
func (sc *SpinnerColumn) Render(task TaskSnapshot, c *console.Console, opts console.Options) []segment.Segment {
	if task.Finished {
		s := style.Parse("green")
		if sc.Style != nil {
			s = *sc.Style
		}
		return []segment.Segment{segment.NewText(sc.FinishedText, &s)}
	}

	sc.mu.Lock()
	defer sc.mu.Unlock()

	// Get or create spinner for this task
	spin, ok := sc.spinners[task.ID]
	if !ok {
		name := sc.SpinnerName
		if perName, ok := sc.perTaskNames[task.ID]; ok {
			name = perName
		}
		spin = spinner.New(name)
		if sc.Style != nil {
			spin = spin.WithStyle(*sc.Style)
		}
		spin = spin.WithSpeed(sc.Speed)
		sc.spinners[task.ID] = spin
	}

	return spin.Render(task.CurrentTime)
}

// MaxRefresh implements Column.
func (sc *SpinnerColumn) MaxRefresh() time.Duration {
	return sc.maxRefresh
}

// DownloadColumn displays download progress in bytes.
type DownloadColumn struct {
	BinaryUnits bool // Use KiB/MiB/GiB instead of KB/MB/GB
	maxRefresh  time.Duration
}

// NewDownloadColumn creates a new download column.
func NewDownloadColumn(binaryUnits bool) *DownloadColumn {
	return &DownloadColumn{BinaryUnits: binaryUnits}
}

// Render implements Column.
func (dc *DownloadColumn) Render(task TaskSnapshot, c *console.Console, opts console.Options) []segment.Segment {
	completed := formatFileSize(task.Completed, dc.BinaryUnits)
	var text string

	if task.Total != nil {
		total := formatFileSize(*task.Total, dc.BinaryUnits)
		text = fmt.Sprintf("%s/%s", completed, total)
	} else {
		text = completed
	}

	s := style.Parse("green")
	return []segment.Segment{segment.NewText(text, &s)}
}

// MaxRefresh implements Column.
func (dc *DownloadColumn) MaxRefresh() time.Duration {
	return dc.maxRefresh
}

// TransferSpeedColumn displays transfer speed in bytes/second.
type TransferSpeedColumn struct {
	BinaryUnits bool
	maxRefresh  time.Duration
}

// NewTransferSpeedColumn creates a new transfer speed column.
func NewTransferSpeedColumn(binaryUnits bool) *TransferSpeedColumn {
	return &TransferSpeedColumn{BinaryUnits: binaryUnits}
}

// Render implements Column.
func (tsc *TransferSpeedColumn) Render(task TaskSnapshot, c *console.Console, opts console.Options) []segment.Segment {
	speed := task.GetSpeed()
	var text string

	if speed != nil && *speed > 0 {
		text = formatFileSize(*speed, tsc.BinaryUnits) + "/s"
	} else {
		text = "---"
	}

	s := style.Parse("red")
	return []segment.Segment{segment.NewText(text, &s)}
}

// MaxRefresh implements Column.
func (tsc *TransferSpeedColumn) MaxRefresh() time.Duration {
	return tsc.maxRefresh
}

// FileSizeColumn displays the completed amount as a file size.
type FileSizeColumn struct {
	BinaryUnits bool
	maxRefresh  time.Duration
}

// NewFileSizeColumn creates a new file size column.
func NewFileSizeColumn(binaryUnits bool) *FileSizeColumn {
	return &FileSizeColumn{BinaryUnits: binaryUnits}
}

// Render implements Column.
func (fsc *FileSizeColumn) Render(task TaskSnapshot, c *console.Console, opts console.Options) []segment.Segment {
	text := formatFileSize(task.Completed, fsc.BinaryUnits)
	s := style.Parse("green")
	return []segment.Segment{segment.NewText(text, &s)}
}

// MaxRefresh implements Column.
func (fsc *FileSizeColumn) MaxRefresh() time.Duration {
	return fsc.maxRefresh
}

// TotalFileSizeColumn displays the total as a file size.
type TotalFileSizeColumn struct {
	BinaryUnits bool
	maxRefresh  time.Duration
}

// NewTotalFileSizeColumn creates a new total file size column.
func NewTotalFileSizeColumn(binaryUnits bool) *TotalFileSizeColumn {
	return &TotalFileSizeColumn{BinaryUnits: binaryUnits}
}

// Render implements Column.
func (tfsc *TotalFileSizeColumn) Render(task TaskSnapshot, c *console.Console, opts console.Options) []segment.Segment {
	var text string
	if task.Total != nil {
		text = formatFileSize(*task.Total, tfsc.BinaryUnits)
	} else {
		text = "?"
	}
	s := style.Parse("green")
	return []segment.Segment{segment.NewText(text, &s)}
}

// MaxRefresh implements Column.
func (tfsc *TotalFileSizeColumn) MaxRefresh() time.Duration {
	return tfsc.maxRefresh
}

// RenderableColumn displays a static renderable.
type RenderableColumn struct {
	Renderable console.Renderable
	maxRefresh time.Duration
}

// NewRenderableColumn creates a new renderable column.
func NewRenderableColumn(r console.Renderable) *RenderableColumn {
	return &RenderableColumn{Renderable: r}
}

// Render implements Column.
func (rc *RenderableColumn) Render(task TaskSnapshot, c *console.Console, opts console.Options) []segment.Segment {
	if rc.Renderable == nil {
		return nil
	}
	return rc.Renderable.Render(c, opts)
}

// MaxRefresh implements Column.
func (rc *RenderableColumn) MaxRefresh() time.Duration {
	return rc.maxRefresh
}

// File size formatting helpers

var decimalUnits = []string{"B", "KB", "MB", "GB", "TB", "PB", "EB"}
var binaryUnits = []string{"B", "KiB", "MiB", "GiB", "TiB", "PiB", "EiB"}

func formatFileSize(size float64, binary bool) string {
	if size < 0 {
		size = 0
	}

	units := decimalUnits
	base := 1000.0
	if binary {
		units = binaryUnits
		base = 1024.0
	}

	unitIndex := 0
	for size >= base && unitIndex < len(units)-1 {
		size /= base
		unitIndex++
	}

	if unitIndex == 0 {
		return fmt.Sprintf("%.0f %s", size, units[unitIndex])
	}
	return fmt.Sprintf("%.1f %s", size, units[unitIndex])
}
