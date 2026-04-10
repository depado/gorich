package progress

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/depado/gorich/console"
	"github.com/depado/gorich/live"
	"github.com/depado/gorich/segment"
)

// Progress manages multiple progress tasks.
type Progress struct {
	mu                  sync.Mutex
	console             *console.Console
	columns             []Column
	tasks               map[TaskID]*Task
	taskOrder           []TaskID // Maintains insertion order
	nextTaskID          TaskID
	live                *live.Live
	speedEstimatePeriod float64
	refreshRate         float64
	transient           bool
	disable             bool
	expand              bool
	started             bool
	getTime             func() float64
}

// Option configures a Progress.
type Option func(*Progress)

// WithConsole sets the console to use.
func WithConsole(c *console.Console) Option {
	return func(p *Progress) {
		p.console = c
	}
}

// WithColumns sets the columns to display.
func WithColumns(columns ...Column) Option {
	return func(p *Progress) {
		p.columns = columns
	}
}

// WithSpeedEstimatePeriod sets the window for speed estimation (default 30s).
func WithSpeedEstimatePeriod(seconds float64) Option {
	return func(p *Progress) {
		p.speedEstimatePeriod = seconds
	}
}

// WithRefreshRate sets the refresh rate in Hz (default 10).
func WithRefreshRate(hz float64) Option {
	return func(p *Progress) {
		p.refreshRate = hz
	}
}

// WithTransient clears the progress display when stopped.
func WithTransient(transient bool) Option {
	return func(p *Progress) {
		p.transient = transient
	}
}

// WithDisable disables all output (useful for CI).
func WithDisable(disable bool) Option {
	return func(p *Progress) {
		p.disable = disable
	}
}

// WithExpand expands the display to fill terminal width.
func WithExpand(expand bool) Option {
	return func(p *Progress) {
		p.expand = expand
	}
}

// New creates a new Progress with the given options.
func New(opts ...Option) *Progress {
	p := &Progress{
		tasks:               make(map[TaskID]*Task),
		speedEstimatePeriod: 30.0,
		refreshRate:         10.0,
		getTime:             defaultGetTime,
	}

	for _, opt := range opts {
		opt(p)
	}

	if p.console == nil {
		p.console = console.New()
	}

	if len(p.columns) == 0 {
		p.columns = DefaultColumns()
	}

	return p
}

func defaultGetTime() float64 {
	return float64(time.Now().UnixNano()) / 1e9
}

// Start begins the progress display.
func (p *Progress) Start(ctx context.Context) {
	p.mu.Lock()

	if p.started || p.disable {
		p.mu.Unlock()
		return
	}

	p.started = true

	// Create Live display
	p.live = live.New(
		p.console,
		nil, // We'll use GetRenderable callback
		live.WithTransient(p.transient),
		live.WithRefreshRate(p.refreshRate),
		live.WithGetRenderable(p.getRenderable),
	)

	// Release lock before starting live (it will call getRenderable which needs the lock)
	l := p.live
	p.mu.Unlock()

	l.Start(ctx)
}

// Stop ends the progress display.
func (p *Progress) Stop() {
	p.mu.Lock()

	if !p.started {
		p.mu.Unlock()
		return
	}

	p.started = false
	l := p.live
	p.live = nil
	p.mu.Unlock()

	if l != nil {
		l.Stop()
	}
}

// AddTask adds a new task and returns its ID.
func (p *Progress) AddTask(description string, total *float64, opts ...TaskOption) TaskID {
	p.mu.Lock()
	defer p.mu.Unlock()

	cfg := TaskConfig{
		Description: description,
		Total:       total,
		Visible:     true,
		Start:       true,
	}

	for _, opt := range opts {
		opt(&cfg)
	}

	id := p.nextTaskID
	p.nextTaskID++

	task := NewTask(id, cfg, p.getTime, p.speedEstimatePeriod)
	p.tasks[id] = task
	p.taskOrder = append(p.taskOrder, id)

	return id
}

// TaskOption configures task creation.
type TaskOption func(*TaskConfig)

// TaskWithCompleted sets the initial completed count.
func TaskWithCompleted(completed float64) TaskOption {
	return func(cfg *TaskConfig) {
		cfg.Completed = completed
	}
}

// TaskWithVisible sets task visibility.
func TaskWithVisible(visible bool) TaskOption {
	return func(cfg *TaskConfig) {
		cfg.Visible = visible
	}
}

// TaskWithStart sets whether to start timing immediately.
func TaskWithStart(start bool) TaskOption {
	return func(cfg *TaskConfig) {
		cfg.Start = start
	}
}

// TaskWithFields sets custom fields.
func TaskWithFields(fields map[string]any) TaskOption {
	return func(cfg *TaskConfig) {
		cfg.Fields = fields
	}
}

// Advance increments a task's progress.
func (p *Progress) Advance(taskID TaskID, amount float64) {
	p.mu.Lock()
	task, ok := p.tasks[taskID]
	p.mu.Unlock()

	if !ok {
		return
	}

	task.Advance(amount)
}

// Update updates a task's state.
func (p *Progress) Update(taskID TaskID, cfg TaskUpdateConfig) {
	p.mu.Lock()
	task, ok := p.tasks[taskID]
	p.mu.Unlock()

	if !ok {
		return
	}

	task.Update(cfg)
}

// RemoveTask removes a task from the display.
func (p *Progress) RemoveTask(taskID TaskID) {
	p.mu.Lock()
	defer p.mu.Unlock()

	delete(p.tasks, taskID)

	// Remove from order slice
	for i, id := range p.taskOrder {
		if id == taskID {
			p.taskOrder = append(p.taskOrder[:i], p.taskOrder[i+1:]...)
			break
		}
	}
}

// StartTask begins timing a task.
func (p *Progress) StartTask(taskID TaskID) {
	p.mu.Lock()
	task, ok := p.tasks[taskID]
	p.mu.Unlock()

	if ok {
		task.Start()
	}
}

// StopTask pauses timing a task.
func (p *Progress) StopTask(taskID TaskID) {
	p.mu.Lock()
	task, ok := p.tasks[taskID]
	p.mu.Unlock()

	if ok {
		task.Stop()
	}
}

// ResetTask resets a task to initial state.
func (p *Progress) ResetTask(taskID TaskID, start bool) {
	p.mu.Lock()
	task, ok := p.tasks[taskID]
	p.mu.Unlock()

	if ok {
		task.Reset(start)
	}
}

// Refresh forces an immediate refresh.
func (p *Progress) Refresh() {
	p.mu.Lock()
	l := p.live
	p.mu.Unlock()

	if l != nil {
		l.Refresh()
	}
}

// Finished returns true if all tasks are finished.
func (p *Progress) Finished() bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, task := range p.tasks {
		snap := task.Snapshot()
		if !snap.Finished {
			return false
		}
	}

	return len(p.tasks) > 0
}

// getRenderable returns the current renderable for Live.
func (p *Progress) getRenderable() console.Renderable {
	p.mu.Lock()
	defer p.mu.Unlock()

	return p.makeRenderable()
}

// makeRenderable creates the renderable (must hold lock).
func (p *Progress) makeRenderable() console.Renderable {
	// Get visible task snapshots in order
	var snapshots []TaskSnapshot
	for _, id := range p.taskOrder {
		if task, ok := p.tasks[id]; ok {
			snap := task.Snapshot()
			if snap.Visible {
				snapshots = append(snapshots, snap)
			}
		}
	}

	if len(snapshots) == 0 {
		return nil
	}

	return &progressRenderable{
		columns:   p.columns,
		snapshots: snapshots,
	}
}

// progressRenderable renders all tasks.
type progressRenderable struct {
	columns   []Column
	snapshots []TaskSnapshot
}

// Render implements console.Renderable.
func (pr *progressRenderable) Render(c *console.Console, opts console.Options) []segment.Segment {
	var allSegments []segment.Segment

	for i, snap := range pr.snapshots {
		// Render each column for this task
		var lineSegments []segment.Segment

		for j, col := range pr.columns {
			colSegments := col.Render(snap, c, opts)
			lineSegments = append(lineSegments, colSegments...)

			// Add spacing between columns (except after last)
			if j < len(pr.columns)-1 {
				lineSegments = append(lineSegments, segment.Segment{Text: " "})
			}
		}

		allSegments = append(allSegments, lineSegments...)

		// Add newline between tasks (except after last)
		if i < len(pr.snapshots)-1 {
			allSegments = append(allSegments, segment.Segment{Text: "\n"})
		}
	}

	return allSegments
}

// Track iterates over items while showing progress.
func Track[T any](items []T, description string, opts ...Option) func(func(T) bool) {
	total := float64(len(items))
	p := New(opts...)

	return func(yield func(T) bool) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		p.Start(ctx)
		defer p.Stop()

		taskID := p.AddTask(description, &total)

		for _, item := range items {
			if !yield(item) {
				return
			}
			p.Advance(taskID, 1)
		}
	}
}

// Simple string builder for progress output when not using Live
func (p *Progress) String() string {
	p.mu.Lock()
	defer p.mu.Unlock()

	var b strings.Builder
	opts := p.console.Options()

	for _, id := range p.taskOrder {
		if task, ok := p.tasks[id]; ok {
			snap := task.Snapshot()
			if !snap.Visible {
				continue
			}

			for j, col := range p.columns {
				segments := col.Render(snap, p.console, opts)
				for _, seg := range segments {
					b.WriteString(seg.Text)
				}
				if j < len(p.columns)-1 {
					b.WriteString(" ")
				}
			}
			b.WriteString("\n")
		}
	}

	return b.String()
}
