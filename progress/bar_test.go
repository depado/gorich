package progress

import (
	"strings"
	"testing"

	"github.com/depado/gorich/console"
	"github.com/depado/gorich/segment"
)

func TestBarRenderDeterminate(t *testing.T) {
	c := console.New(console.WithNoColor(true), console.WithForceTerminal(true))

	tests := []struct {
		name        string
		total       *float64
		completed   float64
		width       *int
		finished    bool
		wantTextLen int
		wantText    string
	}{
		{
			name:        "0% complete",
			total:       new(10.0),
			completed:   0,
			width:       new(10),
			wantTextLen: 10,
			wantText:    "          ",
		},
		{
			name:        "50% complete",
			total:       new(10.0),
			completed:   5,
			width:       new(10),
			wantTextLen: 10,
			wantText:    "━━━━━     ",
		},
		{
			name:        "100% complete",
			total:       new(10.0),
			completed:   10,
			width:       new(10),
			finished:    true,
			wantTextLen: 10,
			wantText:    "━━━━━━━━━━",
		},
		{
			name:        "33% at width 12",
			total:       new(12.0),
			completed:   4,
			width:       new(12),
			wantTextLen: 12,
			wantText:    "━━━━        ",
		},
		{
			name:        "half char boundary",
			total:       new(4.0),
			completed:   1,
			width:       new(3),
			wantTextLen: 3,
			wantText:    "╸  ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pb := &ProgressBar{
				Total:     tt.total,
				Completed: tt.completed,
				Width:     tt.width,
				Finished:  tt.finished,
			}
			segs := pb.Render(c, c.Options())

			if segs == nil {
				t.Fatal("expected non-nil segments")
			}

			totalLen := segment.TotalCellLength(segs)
			if totalLen != tt.wantTextLen {
				t.Errorf("TotalCellLength = %d, want %d", totalLen, tt.wantTextLen)
			}

			var b strings.Builder
			for _, s := range segs {
				b.WriteString(s.Text)
			}
			got := b.String()
			if got != tt.wantText {
				t.Errorf("text = %q, want %q", got, tt.wantText)
			}
		})
	}
}

func TestBarRenderPulse(t *testing.T) {
	c := console.New(console.WithNoColor(true), console.WithForceTerminal(true))

	tests := []struct {
		name      string
		total     *float64
		width     *int
		pulse     bool
		wantWidth int
	}{
		{
			name:      "default width",
			total:     nil,
			wantWidth: 40,
		},
		{
			name:      "explicit width",
			total:     nil,
			width:     new(5),
			wantWidth: 5,
		},
		{
			name:      "pulse forced",
			total:     new(100.0),
			pulse:     true,
			width:     new(8),
			wantWidth: 8,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pb := &ProgressBar{
				Total: tt.total,
				Width: tt.width,
				Pulse: tt.pulse,
			}
			segs := pb.Render(c, c.Options())

			if segs == nil {
				t.Fatal("expected non-nil segments")
			}

			if len(segs) < 1 {
				t.Error("expected at least 1 segment")
			}

			totalLen := segment.TotalCellLength(segs)
			if totalLen != tt.wantWidth {
				t.Errorf("TotalCellLength = %d, want %d", totalLen, tt.wantWidth)
			}

			for i, s := range segs {
				if s.Style == nil {
					t.Errorf("segment %d has nil style", i)
				}
			}
		})
	}
}

func TestBarFinishedStyle(t *testing.T) {
	c := console.New(console.WithNoColor(true), console.WithForceTerminal(true))

	pb := &ProgressBar{
		Total:     new(10.0),
		Completed: 10.0,
		Width:     new(5),
		Finished:  true,
	}
	segs := pb.Render(c, c.Options())

	if len(segs) == 0 {
		t.Fatal("expected non-empty segments")
	}

	for i, s := range segs {
		if s.Style == nil {
			t.Errorf("segment %d has nil style", i)
			continue
		}
		fg := s.Style.Foreground()
		if fg == nil {
			t.Errorf("segment %d has nil foreground", i)
			continue
		}
		r, g, b := fg.RGB()
		if r != 0 || g != 255 || b != 0 {
			t.Errorf("segment %d foreground RGB = (%d,%d,%d), want (0,255,0)", i, r, g, b)
		}
	}
}

func TestBarASCII(t *testing.T) {
	c := console.New(console.WithNoColor(true), console.WithForceTerminal(true))

	pb := &ProgressBar{
		Total:     new(10.0),
		Completed: 5.0,
		Width:     new(5),
		ASCIIOnly: true,
	}
	segs := pb.Render(c, c.Options())

	if len(segs) == 0 {
		t.Fatal("expected non-empty segments")
	}

	// Completed portion uses "-" and empty uses " "; at 50% width=5 that's
	// 2 full + 1 half (also "-" in ASCII) + 2 empty = "---  "
	var b strings.Builder
	for _, s := range segs {
		b.WriteString(s.Text)
	}
	got := b.String()
	if !strings.Contains(got, barFullASCII) {
		t.Errorf("text %q does not contain ASCII dash %q", got, barFullASCII)
	}
	if strings.Contains(got, barFull) {
		t.Errorf("text %q contains unicode full bar %q in ASCII mode", got, barFull)
	}
}

func TestBarMeasure(t *testing.T) {
	c := console.New(console.WithForceTerminal(true))

	t.Run("explicit width", func(t *testing.T) {
		w := 20
		pb := &ProgressBar{Width: &w}
		m := pb.Measure(c, c.Options())
		if m.Minimum != 20 || m.Maximum != 20 {
			t.Errorf("Measure = (%d,%d), want (20,20)", m.Minimum, m.Maximum)
		}
	})

	t.Run("default width", func(t *testing.T) {
		pb := &ProgressBar{}
		m := pb.Measure(c, c.Options())
		if m.Minimum != 40 || m.Maximum != 40 {
			t.Errorf("Measure = (%d,%d), want (40,40)", m.Minimum, m.Maximum)
		}
	})
}
