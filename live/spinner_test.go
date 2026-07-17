package live

import (
	"testing"

	"github.com/depado/gorich/console"
)

func TestActiveSpinnerLifecycle(t *testing.T) {
	c := console.New(
		console.WithNoColor(true),
		console.WithForceTerminal(true),
		console.WithWidth(80),
	)
	a := StartSpinner("working...", WithSpinnerConsole(c), WithSpinnerRefreshRate(50))

	segs := a.Render(c, c.Options())
	if len(segs) == 0 {
		t.Error("expected spinner segments while spinning")
	}

	a.mu.Lock()
	if a.state != stateSpinning {
		t.Error("expected stateSpinning")
	}
	a.mu.Unlock()

	a.Stop()
}

func TestActiveSpinnerFail(t *testing.T) {
	c := console.New(
		console.WithNoColor(true),
		console.WithForceTerminal(true),
		console.WithWidth(80),
	)
	a := StartSpinner("working...", WithSpinnerConsole(c), WithSpinnerRefreshRate(50))

	a.Fail("failed")
	a.mu.Lock()
	if a.state != stateFailure {
		t.Error("expected stateFailure after Fail()")
	}
	a.mu.Unlock()

	a.Stop()
}

func TestActiveSpinnerStop(t *testing.T) {
	c := console.New(
		console.WithNoColor(true),
		console.WithForceTerminal(true),
		console.WithWidth(80),
	)
	a := StartSpinner("working...", WithSpinnerConsole(c), WithSpinnerRefreshRate(50))
	a.Stop()
	segs := a.Render(c, c.Options())
	if segs != nil {
		t.Error("expected nil segments after Stop()")
	}
}

func TestActiveSpinnerUpdate(t *testing.T) {
	c := console.New(
		console.WithNoColor(true),
		console.WithForceTerminal(true),
		console.WithWidth(80),
	)
	a := StartSpinner("working...", WithSpinnerConsole(c), WithSpinnerRefreshRate(50))
	a.Update("still working...")

	segs := a.Render(c, c.Options())
	text := renderSegsToString(segs)
	if !stringsContains(text, "still working...") {
		t.Errorf("expected updated text, got: %s", text)
	}

	a.Stop()
}

func stringsContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
