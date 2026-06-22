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
			total:       ptrFloat(10.0),
			completed:   0,
			width:       ptrInt(10),
			wantTextLen: 10,
			wantText:    "          ",
		},
		{
			name:        "50% complete",
			total:       ptrFloat(10.0),
			completed:   5,
			width:       ptrInt(10),
			wantTextLen: 10,
			wantText:    "━━━━━     ",
		},
		{
			name:        "100% complete",
			total:       ptrFloat(10.0),
			completed:   10,
			width:       ptrInt(10),
			finished:    true,
			wantTextLen: 10,
			wantText:    "━━━━━━━━━━",
		},
		{
			name:        "33% at width 12",
			total:       ptrFloat(12.0),
			completed:   4,
			width:       ptrInt(12),
			wantTextLen: 12,
			wantText:    "━━━━        ",
		},
		{
			name:        "half char boundary",
			total:       ptrFloat(4.0),
			completed:   1,
			width:       ptrInt(3),
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
		wantCount int
	}{
		{
			name:      "default width",
			total:     nil,
			wantCount: 40,
		},
		{
			name:      "explicit width",
			total:     nil,
			width:     ptrInt(5),
			wantCount: 5,
		},
		{
			name:      "pulse forced",
			total:     ptrFloat(100.0),
			pulse:     true,
			width:     ptrInt(8),
			wantCount: 8,
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

			if len(segs) != tt.wantCount {
				t.Errorf("segment count = %d, want %d", len(segs), tt.wantCount)
			}

			totalLen := segment.TotalCellLength(segs)
			if totalLen != tt.wantCount {
				t.Errorf("TotalCellLength = %d, want %d", totalLen, tt.wantCount)
			}

			for i, s := range segs {
				if s.Style == nil {
					t.Errorf("segment %d has nil style", i)
				}
				if s.Text != barFull {
					t.Errorf("segment %d text = %q, want %q", i, s.Text, barFull)
				}
			}
		})
	}
}

func TestBarFinishedStyle(t *testing.T) {
	c := console.New(console.WithNoColor(true), console.WithForceTerminal(true))

	pb := &ProgressBar{
		Total:     ptrFloat(10.0),
		Completed: 10.0,
		Width:     ptrInt(5),
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
		Total:     ptrFloat(10.0),
		Completed: 5.0,
		Width:     ptrInt(5),
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

func ptrInt(i int) *int {
	return &i
}

func ptrFloat(f float64) *float64 {
	return &f
}
