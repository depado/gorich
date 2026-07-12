package main

import (
	"context"
	"os/exec"
	"sync"
	"time"

	"github.com/depado/gorich/console"
	"github.com/depado/gorich/live"
)

func main() {
	c := console.New()
	display := live.NewBlockDisplay(
		live.WithBlockMaxLines(3),
		live.WithBlockSpinnerName("dots"),
		live.WithBlockPrefix("[dim]│ [/]"),
	)

	// Wrap with Live so the spinner animates and lines refresh in place.
	l := live.New(c, display, live.WithAutoRefresh(true), live.WithRefreshRate(15))
	ctx := context.Background()
	l.Start(ctx)

	// Run commands in parallel, streaming their output to per-block writers.
	var wg sync.WaitGroup
	repos := []struct {
		name string
		expr string
	}{
		{"alpha", "echo hi && sleep 1 && echo there && sleep 1 && echo done"},
		{"beta", "echo start && sleep 1 && echo middle && sleep 0.5 && echo more"},
		{"gamma", "echo one && echo two && sleep 2 && echo three && echo four"},
		{"delta", "echo failing && echo badly && sleep 1 && exit 2"},
	}
	for _, r := range repos {
		wg.Add(1)
		go func(r struct {
			name string
			expr string
		}) {
			defer wg.Done()
			idx := display.Start(r.name)
			out := display.NewWriter(idx)

			cmd := exec.CommandContext(ctx, "sh", "-c", r.expr)
			cmd.Stdout = out
			cmd.Stderr = out
			err := cmd.Run()
			if err != nil {
				if exitErr, ok := err.(*exec.ExitError); ok {
					display.Finish(idx, exitErr.ExitCode())
				} else {
					display.Finish(idx, 1)
				}
			} else {
				display.Finish(idx, 0)
			}
		}(r)
	}

	wg.Wait()

	// Give the final state a moment to render before stopping the live display.
	time.Sleep(300 * time.Millisecond)

	// Stop the live display so cursor is restored and the final render sticks.
	l.Stop()
}
