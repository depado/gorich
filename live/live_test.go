package live

import (
	"context"
	"testing"
	"time"

	"github.com/depado/gorich/console"
)

func TestLiveStartStop(t *testing.T) {
	c := console.New(
		console.WithNoColor(true),
		console.WithForceTerminal(true),
		console.WithWidth(80),
	)
	l := New(c, &dummyRenderable{}, WithAutoRefresh(false))
	ctx := context.Background()
	l.Start(ctx)

	if !l.isStarted() {
		t.Error("expected started after Start()")
	}

	l.Stop()
	if l.isStarted() {
		t.Error("expected not started after Stop()")
	}
}

func TestLiveDoubleStartNoOp(t *testing.T) {
	c := console.New(
		console.WithNoColor(true),
		console.WithForceTerminal(true),
		console.WithWidth(80),
	)
	l := New(c, &dummyRenderable{}, WithAutoRefresh(false))
	ctx := context.Background()
	l.Start(ctx)
	l.Start(ctx) // should not panic

	l.Stop()
}

func TestLiveAutoRefresh(t *testing.T) {
	c := console.New(
		console.WithNoColor(true),
		console.WithForceTerminal(true),
		console.WithWidth(80),
	)
	l := New(c, &dummyRenderable{}, WithAutoRefresh(true), WithRefreshRate(50))
	ctx := context.Background()
	l.Start(ctx)

	time.Sleep(100 * time.Millisecond)

	l.Stop()
}

func TestLiveGetRenderable(t *testing.T) {
	c := console.New(
		console.WithNoColor(true),
		console.WithForceTerminal(true),
		console.WithWidth(80),
	)
	called := false
	l := New(c, nil, WithAutoRefresh(false), WithGetRenderable(func() console.Renderable {
		called = true
		return &dummyRenderable{}
	}))
	ctx := context.Background()
	l.Start(ctx)
	l.Stop()

	if !called {
		t.Error("getRenderable callback should have been called")
	}
}

func TestLiveTransient(t *testing.T) {
	c := console.New(
		console.WithNoColor(true),
		console.WithForceTerminal(true),
		console.WithWidth(80),
	)
	l := New(c, &dummyRenderable{}, WithAutoRefresh(false), WithTransient(true))
	ctx := context.Background()
	l.Start(ctx)
	l.Stop()
}
