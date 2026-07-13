package live

import (
	"context"
	"sync"
	"time"

	"github.com/depado/gorich/console"
	"github.com/depado/gorich/segment"
	"github.com/depado/gorich/spinner"
	"github.com/depado/gorich/style"
)

type activeState int

const (
	stateSpinning activeState = iota
	stateSuccess
	stateFailure
	stateStopped
)

type ActiveSpinner struct {
	mu          sync.Mutex
	live        *Live
	console     *console.Console
	spin        *spinner.Spinner
	text        string
	state       activeState
	started     bool
	name        string
	style       *style.Style
	speed       float64
	refreshRate float64
}

type ActiveSpinnerOption func(*ActiveSpinner)

func WithSpinnerName(name string) ActiveSpinnerOption {
	return func(a *ActiveSpinner) {
		a.name = name
	}
}

func WithSpinnerStyle(s style.Style) ActiveSpinnerOption {
	return func(a *ActiveSpinner) {
		a.style = &s
	}
}

func WithSpinnerSpeed(speed float64) ActiveSpinnerOption {
	return func(a *ActiveSpinner) {
		a.speed = speed
	}
}

func WithSpinnerRefreshRate(hz float64) ActiveSpinnerOption {
	return func(a *ActiveSpinner) {
		a.refreshRate = hz
	}
}

func WithSpinnerConsole(c *console.Console) ActiveSpinnerOption {
	return func(a *ActiveSpinner) {
		a.console = c
	}
}

func StartSpinner(text string, opts ...ActiveSpinnerOption) *ActiveSpinner {
	a := &ActiveSpinner{
		text:        text,
		state:       stateSpinning,
		name:        "dots",
		speed:       1.0,
		refreshRate: 10.0,
	}
	for _, opt := range opts {
		opt(a)
	}
	if a.console == nil {
		a.console = console.New()
	}

	a.spin = spinner.New(a.name).WithSpeed(a.speed)
	if a.style != nil {
		a.spin = a.spin.WithStyle(*a.style)
	}

	a.live = New(
		a.console,
		nil,
		WithRefreshRate(a.refreshRate),
		WithGetRenderable(a.getRenderable),
	)

	a.started = true
	a.live.Start(context.Background())
	return a
}

func (a *ActiveSpinner) getRenderable() console.Renderable {
	return a
}

func (a *ActiveSpinner) Render(c *console.Console, opts console.Options) []segment.Segment {
	a.mu.Lock()
	spin := a.spin
	text := a.text
	state := a.state
	a.mu.Unlock()

	switch state {
	case stateStopped:
		return nil
	case stateSuccess:
		return []segment.Segment{
			segment.NewText("✓", &style.Green),
			segment.NewText(" "+text, nil),
		}
	case stateFailure:
		return []segment.Segment{
			segment.NewText("✗", &style.Red),
			segment.NewText(" "+text, nil),
		}
	default:
		now := float64(time.Now().UnixNano()) / 1e9
		segs := spin.Render(now)
		segs = append(segs, segment.NewText(" "+text, nil))
		return segs
	}
}

func (a *ActiveSpinner) Update(text string) {
	a.mu.Lock()
	a.text = text
	a.mu.Unlock()
}

func (a *ActiveSpinner) Succeed(text ...string) {
	a.mu.Lock()
	if !a.started {
		a.mu.Unlock()
		return
	}
	a.state = stateSuccess
	if len(text) > 0 {
		a.text = text[0]
	}
	a.mu.Unlock()

	time.Sleep(time.Duration(float64(time.Second)/a.refreshRate) * 2)
	a.stop()
}

func (a *ActiveSpinner) Fail(text ...string) {
	a.mu.Lock()
	if !a.started {
		a.mu.Unlock()
		return
	}
	a.state = stateFailure
	if len(text) > 0 {
		a.text = text[0]
	}
	a.mu.Unlock()

	time.Sleep(time.Duration(float64(time.Second)/a.refreshRate) * 2)
	a.stop()
}

func (a *ActiveSpinner) Stop() {
	a.mu.Lock()
	if !a.started {
		a.mu.Unlock()
		return
	}
	a.state = stateStopped
	a.mu.Unlock()
	a.stop()
	a.console.WriteControl(segment.CursorUp(1))
	a.console.WriteControl(segment.CarriageReturn())
}

func (a *ActiveSpinner) stop() {
	a.mu.Lock()
	if !a.started {
		a.mu.Unlock()
		return
	}
	a.started = false
	l := a.live
	a.mu.Unlock()

	if l != nil {
		l.Stop()
	}
}
