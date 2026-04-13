package live

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/depado/gorich/console"
	"github.com/depado/gorich/segment"
)

// Live provides an auto-refreshing terminal display.
type Live struct {
	mu            sync.Mutex
	console       *console.Console
	renderable    console.Renderable
	getRenderable func() console.Renderable // Optional callback to get renderable
	liveRender    *LiveRender
	started       bool
	transient     bool
	autoRefresh   bool
	refreshPeriod time.Duration
	vertOverflow  VerticalOverflow
	stopCh        chan struct{}
	doneCh        chan struct{}
}

// Option configures a Live display.
type Option func(*Live)

// WithTransient makes the live display disappear when stopped.
func WithTransient(transient bool) Option {
	return func(l *Live) {
		l.transient = transient
	}
}

// WithAutoRefresh enables automatic refresh at the given rate.
func WithAutoRefresh(enabled bool) Option {
	return func(l *Live) {
		l.autoRefresh = enabled
	}
}

// WithRefreshRate sets the refresh rate (default 10Hz).
func WithRefreshRate(rate float64) Option {
	return func(l *Live) {
		if rate > 0 {
			l.refreshPeriod = time.Duration(float64(time.Second) / rate)
		}
	}
}

// WithVerticalOverflow sets how vertical overflow is handled.
func WithVerticalOverflow(overflow VerticalOverflow) Option {
	return func(l *Live) {
		l.vertOverflow = overflow
	}
}

// WithGetRenderable sets a callback to get the renderable on each refresh.
func WithGetRenderable(fn func() console.Renderable) Option {
	return func(l *Live) {
		l.getRenderable = fn
	}
}

// New creates a new Live display.
func New(c *console.Console, renderable console.Renderable, opts ...Option) *Live {
	l := &Live{
		console:       c,
		renderable:    renderable,
		autoRefresh:   true,
		refreshPeriod: 100 * time.Millisecond, // 10Hz default
		vertOverflow:  OverflowVisible,
	}

	for _, opt := range opts {
		opt(l)
	}

	l.liveRender = NewLiveRender(renderable, l.vertOverflow)

	return l
}

// Start begins the live display.
func (l *Live) Start(ctx context.Context) {
	l.mu.Lock()
	if l.started {
		l.mu.Unlock()
		return
	}
	l.started = true
	l.stopCh = make(chan struct{})
	l.doneCh = make(chan struct{})
	l.mu.Unlock()

	// Hide cursor
	if l.console.IsTerminal() {
		l.console.WriteControl(segment.HideCursor())
	}

	// Register as render hook
	l.console.PushRenderHook(l)

	// Initial render
	l.refresh()

	// Start refresh goroutine if auto-refresh is enabled
	if l.autoRefresh {
		go l.refreshLoop(ctx)
	}
}

// Stop ends the live display.
func (l *Live) Stop() {
	l.mu.Lock()
	if !l.started {
		l.mu.Unlock()
		return
	}
	l.started = false
	l.mu.Unlock()

	// Signal refresh goroutine to stop
	close(l.stopCh)

	// Wait for goroutine to finish
	if l.autoRefresh {
		<-l.doneCh
	}

	// Final refresh to show the last state
	l.mu.Lock()
	l.refresh()
	l.mu.Unlock()

	// Unregister render hook
	l.console.PopRenderHook()

	// Show cursor
	if l.console.IsTerminal() {
		l.console.WriteControl(segment.ShowCursor())
	}

	// Handle transient mode
	if l.transient && l.console.IsTerminal() {
		l.console.WriteControl(l.liveRender.RestoreCursor())
	} else {
		// Move to next line after the live display
		l.console.WriteString("\n") //nolint:errcheck // terminal output is fire-and-forget
	}
}

// Update changes the renderable being displayed.
func (l *Live) Update(r console.Renderable) {
	l.mu.Lock()
	l.renderable = r
	l.liveRender.SetRenderable(r)
	l.mu.Unlock()
}

// Refresh forces an immediate refresh.
func (l *Live) Refresh() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.refresh()
}

func (l *Live) refresh() {
	if !l.console.IsTerminal() {
		return
	}

	// Get renderable (either stored or from callback)
	renderable := l.renderable
	if l.getRenderable != nil {
		renderable = l.getRenderable()
		l.liveRender.SetRenderable(renderable)
	}

	if renderable == nil {
		return
	}

	// Build everything into a single buffer for flicker-free output
	var buf strings.Builder

	// Position cursor to overwrite previous content
	posCtrl := l.liveRender.PositionCursor()
	if len(posCtrl.Codes) > 0 {
		buf.WriteString(posCtrl.Render())
	}

	// Render the content
	opts := l.console.Options()
	segments := l.liveRender.Render(l.console, opts)

	// Write all segments to buffer
	colorSys := l.console.ColorSystem()
	for _, seg := range segments {
		buf.WriteString(seg.Render(colorSys))
	}

	// Single atomic write to terminal
	l.console.WriteString(buf.String()) //nolint:errcheck // terminal output is fire-and-forget
}

func (l *Live) refreshLoop(ctx context.Context) {
	defer close(l.doneCh)

	ticker := time.NewTicker(l.refreshPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			l.Refresh()
		case <-ctx.Done():
			return
		case <-l.stopCh:
			return
		}
	}
}

// ProcessRenderables implements console.RenderHook.
// This intercepts all console.Render calls while Live is active.
//
// Note: We read console properties BEFORE acquiring l.mu to avoid deadlock.
// console.Render() holds c.mu when calling this method, so calling
// l.console.IsTerminal() while holding l.mu would try to acquire c.mu again.
func (l *Live) ProcessRenderables(renderables []console.Renderable) []console.Renderable {
	// Read console properties before acquiring l.mu to avoid deadlock
	// (caller console.Render already holds c.mu)
	isTerminal := l.console.IsTerminal()

	l.mu.Lock()
	defer l.mu.Unlock()

	if !isTerminal {
		return renderables
	}

	// Prepend cursor repositioning to clear previous live content
	posCtrl := l.liveRender.PositionCursor()

	// Result: [position_cursor, ...user_renderables, live_render]
	result := make([]console.Renderable, 0, len(renderables)+2)

	if len(posCtrl.Codes) > 0 {
		result = append(result, console.ControlRenderable{Control: posCtrl})
	}

	result = append(result, renderables...)
	result = append(result, l.liveRender)

	return result
}

// IsStarted returns whether the live display is running.
func (l *Live) IsStarted() bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.started
}
