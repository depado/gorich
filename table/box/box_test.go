package box

import (
	"strings"
	"testing"

	"github.com/depado/gorich/console"
)

func TestNewBox(t *testing.T) {
	b := SQUARE
	if b.topLeft != '┌' {
		t.Errorf("SQUARE topLeft = %c, want ┌", b.topLeft)
	}
	if b.bottomRight != '┘' {
		t.Errorf("SQUARE bottomRight = %c, want ┘", b.bottomRight)
	}
}

func TestBoxString(t *testing.T) {
	b := ASCII
	s := b.String()
	if !strings.Contains(s, "+--+") {
		t.Errorf("ASCII.String() should contain '+--+', got %q", s)
	}
}

func TestGetTop(t *testing.T) {
	tests := []struct {
		name   string
		box    *Box
		widths []int
		want   string
	}{
		{
			name:   "ASCII single column",
			box:    ASCII,
			widths: []int{5},
			want:   "+-----+",
		},
		{
			name:   "ASCII two columns",
			box:    ASCII,
			widths: []int{3, 4},
			want:   "+--------+",
		},
		{
			name:   "ASCII2 two columns",
			box:    ASCII2,
			widths: []int{3, 4},
			want:   "+---+----+",
		},
		{
			name:   "SQUARE two columns",
			box:    SQUARE,
			widths: []int{3, 4},
			want:   "┌───┬────┐",
		},
		{
			name:   "HEAVY_HEAD two columns",
			box:    HEAVY_HEAD,
			widths: []int{3, 4},
			want:   "┏━━━┳━━━━┓",
		},
		{
			name:   "ROUNDED two columns",
			box:    ROUNDED,
			widths: []int{3, 4},
			want:   "╭───┬────╮",
		},
		{
			name:   "DOUBLE two columns",
			box:    DOUBLE,
			widths: []int{3, 4},
			want:   "╔═══╦════╗",
		},
		{
			name:   "MARKDOWN no top edge",
			box:    MARKDOWN,
			widths: []int{3, 4},
			want:   "          ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.box.GetTop(tt.widths)
			if got != tt.want {
				t.Errorf("GetTop(%v) = %q, want %q", tt.widths, got, tt.want)
			}
		})
	}
}

func TestGetBottom(t *testing.T) {
	tests := []struct {
		name   string
		box    *Box
		widths []int
		want   string
	}{
		{
			name:   "ASCII two columns",
			box:    ASCII,
			widths: []int{3, 4},
			want:   "+--------+",
		},
		{
			name:   "ASCII2 two columns",
			box:    ASCII2,
			widths: []int{3, 4},
			want:   "+---+----+",
		},
		{
			name:   "SQUARE two columns",
			box:    SQUARE,
			widths: []int{3, 4},
			want:   "└───┴────┘",
		},
		{
			name:   "HEAVY_HEAD two columns",
			box:    HEAVY_HEAD,
			widths: []int{3, 4},
			want:   "└───┴────┘",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.box.GetBottom(tt.widths)
			if got != tt.want {
				t.Errorf("GetBottom(%v) = %q, want %q", tt.widths, got, tt.want)
			}
		})
	}
}

func TestGetRow(t *testing.T) {
	widths := []int{3, 4}

	t.Run("ASCII2 head row", func(t *testing.T) {
		got := ASCII2.GetRow(widths, "head")
		want := "+---+----+"
		if got != want {
			t.Errorf("ASCII2.GetRow(head) = %q, want %q", got, want)
		}
	})

	t.Run("ASCII head row", func(t *testing.T) {
		got := ASCII.GetRow(widths, "head")
		want := "|---+----|"
		if got != want {
			t.Errorf("ASCII.GetRow(head) = %q, want %q", got, want)
		}
	})

	t.Run("ASCII row divider", func(t *testing.T) {
		got := ASCII.GetRow(widths, "row")
		want := "|---+----|"
		if got != want {
			t.Errorf("ASCII.GetRow(row) = %q, want %q", got, want)
		}
	})

	t.Run("ASCII mid row", func(t *testing.T) {
		got := ASCII.GetRow(widths, "mid")
		want := "|   |    |"
		if got != want {
			t.Errorf("ASCII.GetRow(mid) = %q, want %q", got, want)
		}
	})

	t.Run("ASCII foot row", func(t *testing.T) {
		got := ASCII.GetRow(widths, "foot")
		want := "|---+----|"
		if got != want {
			t.Errorf("ASCII.GetRow(foot) = %q, want %q", got, want)
		}
	})

	t.Run("SQUARE head row", func(t *testing.T) {
		got := SQUARE.GetRow(widths, "head")
		want := "├───┼────┤"
		if got != want {
			t.Errorf("SQUARE.GetRow(head) = %q, want %q", got, want)
		}
	})
}

func TestSubstitute(t *testing.T) {
	opts := console.Options{ASCIIOnly: true}
	if result := HEAVY.Substitute(opts); result != ASCII {
		t.Errorf("HEAVY.Substitute(ASCIIOnly) = %v, want ASCII", result)
	}
	if result := ASCII.Substitute(opts); result != ASCII {
		t.Errorf("ASCII.Substitute(ASCIIOnly) should stay ASCII")
	}

	opts = console.Options{}
	if result := HEAVY.Substitute(opts); result != HEAVY {
		t.Errorf("HEAVY.Substitute(normal) should stay HEAVY, got %v", result)
	}
	if result := ROUNDED.Substitute(opts); result != ROUNDED {
		t.Errorf("ROUNDED.Substitute(normal) should stay ROUNDED, got %v", result)
	}
}

func TestGetPlainHeadedBox(t *testing.T) {
	if result := HEAVY_HEAD.GetPlainHeadedBox(); result != SQUARE {
		t.Errorf("HEAVY_HEAD.GetPlainHeadedBox() should be SQUARE, got %v", result)
	}
	if result := ASCII_DOUBLE_HEAD.GetPlainHeadedBox(); result != ASCII2 {
		t.Errorf("ASCII_DOUBLE_HEAD.GetPlainHeadedBox() should be ASCII2, got %v", result)
	}
	if result := SQUARE.GetPlainHeadedBox(); result != SQUARE {
		t.Errorf("SQUARE.GetPlainHeadedBox() should be SQUARE, got %v", result)
	}
}

func TestNoEdge(t *testing.T) {
	b := SQUARE
	noEdge := b.NoEdge()
	got := noEdge.GetTop([]int{3, 4})
	want := " ───┬──── "
	if got != want {
		t.Errorf("NoEdge SQUARE GetTop = %q, want %q", got, want)
	}
}

func TestAllPredefinedBoxes(t *testing.T) {
	boxes := map[string]*Box{
		"ASCII":               ASCII,
		"ASCII2":              ASCII2,
		"ASCII_DOUBLE_HEAD":   ASCII_DOUBLE_HEAD,
		"SQUARE":              SQUARE,
		"SQUARE_DOUBLE_HEAD":  SQUARE_DOUBLE_HEAD,
		"MINIMAL":             MINIMAL,
		"MINIMAL_HEAVY_HEAD":  MINIMAL_HEAVY_HEAD,
		"MINIMAL_DOUBLE_HEAD": MINIMAL_DOUBLE_HEAD,
		"SIMPLE":              SIMPLE,
		"SIMPLE_HEAD":         SIMPLE_HEAD,
		"SIMPLE_HEAVY":        SIMPLE_HEAVY,
		"HORIZONTALS":         HORIZONTALS,
		"ROUNDED":             ROUNDED,
		"HEAVY":               HEAVY,
		"HEAVY_EDGE":          HEAVY_EDGE,
		"HEAVY_HEAD":          HEAVY_HEAD,
		"DOUBLE":              DOUBLE,
		"DOUBLE_EDGE":         DOUBLE_EDGE,
		"MARKDOWN":            MARKDOWN,
	}

	for name, b := range boxes {
		if b == nil {
			t.Errorf("%s box is nil", name)
			continue
		}
		// Every box should be able to produce top/row/bottom without panicking
		w := []int{3, 4}
		_ = b.GetTop(w)
		_ = b.GetRow(w, "head")
		_ = b.GetRow(w, "row")
		_ = b.GetRow(w, "mid")
		_ = b.GetRow(w, "foot")
		_ = b.GetBottom(w)
	}
}
