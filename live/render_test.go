package live

import (
	"testing"

	"github.com/depado/gorich/console"
	"github.com/depado/gorich/segment"
)

func TestLiveRenderPositionCursor(t *testing.T) {
	lr := newLiveRender(nil, OverflowVisible)

	ctrl := lr.PositionCursor()
	if len(ctrl.Codes) != 0 {
		t.Errorf("expected empty control for height 0, got %d codes", len(ctrl.Codes))
	}

	lr.lastHeight = 3
	ctrl = lr.PositionCursor()
	if len(ctrl.Codes) != 6 {
		t.Errorf("expected 6 codes for height 3, got %d", len(ctrl.Codes))
	}
}

func TestLiveRenderRestoreCursor(t *testing.T) {
	lr := newLiveRender(nil, OverflowVisible)

	ctrl := lr.RestoreCursor()
	if len(ctrl.Codes) != 0 {
		t.Error("expected empty control for height 0")
	}

	lr.lastHeight = 1
	ctrl = lr.RestoreCursor()
	if len(ctrl.Codes) != 2 {
		t.Errorf("expected 2 codes for height 1, got %d", len(ctrl.Codes))
	}
}

func TestLiveRenderReset(t *testing.T) {
	lr := newLiveRender(nil, OverflowVisible)
	lr.lastHeight = 42
	lr.lastWidth = 80
	lr.Reset()
	if lr.lastHeight != 0 || lr.lastWidth != 0 {
		t.Error("Reset() should zero height and width")
	}
}

func TestLiveRenderShape(t *testing.T) {
	lr := newLiveRender(nil, OverflowVisible)
	w, h := lr.Shape()
	if w != 0 || h != 0 {
		t.Error("Shape() should return zeros for unrendered LiveRender")
	}
}

func TestLiveRenderSetRenderable(t *testing.T) {
	lr := newLiveRender(nil, OverflowVisible)
	lr.SetRenderable(&dummyRenderable{})
	if lr.renderable == nil {
		t.Error("SetRenderable should set the renderable")
	}
}

type dummyRenderable struct{}

func (d *dummyRenderable) Render(c *console.Console, opts console.Options) []segment.Segment {
	return []segment.Segment{{Text: "test"}}
}
