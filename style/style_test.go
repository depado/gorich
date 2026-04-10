package style

import (
	"testing"
)

func TestColorParse(t *testing.T) {
	tests := []struct {
		input    string
		wantType ColorType
		wantErr  bool
	}{
		{"red", ColorTypeStandard, false},
		{"bright_blue", ColorTypeStandard, false},
		{"#ff0000", ColorTypeTrueColor, false},
		{"#f00", ColorTypeTrueColor, false},
		{"rgb(255,128,0)", ColorTypeTrueColor, false},
		{"color(123)", ColorType256, false},
		{"default", ColorTypeDefault, false},
		{"", ColorTypeDefault, false},
		{"invalid_color_name", ColorTypeDefault, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			c, err := ParseColor(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseColor(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if !tt.wantErr && c.typ != tt.wantType {
				t.Errorf("ParseColor(%q) type = %v, want %v", tt.input, c.typ, tt.wantType)
			}
		})
	}
}

func TestColorANSICodes(t *testing.T) {
	tests := []struct {
		color      Color
		foreground bool
		want       string
	}{
		{StandardColor(1), true, "31"},  // red foreground
		{StandardColor(1), false, "41"}, // red background
		{StandardColor(9), true, "91"},  // bright red foreground
		{Color256(123), true, "38;5;123"},
		{Color256(123), false, "48;5;123"},
		{TrueColor(255, 128, 0), true, "38;2;255;128;0"},
		{TrueColor(255, 128, 0), false, "48;2;255;128;0"},
		{DefaultColor(), true, "39"},
		{DefaultColor(), false, "49"},
	}

	for _, tt := range tests {
		name := tt.want
		if !tt.foreground {
			name += "_bg"
		}
		t.Run(name, func(t *testing.T) {
			got := tt.color.ANSICodes(tt.foreground)
			if got != tt.want {
				t.Errorf("ANSICodes() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestColorBlend(t *testing.T) {
	black := TrueColor(0, 0, 0)
	white := TrueColor(255, 255, 255)

	// 50% blend should give gray
	gray := black.Blend(white, 0.5)
	r, g, b := gray.RGB()

	if r < 125 || r > 130 || g < 125 || g > 130 || b < 125 || b > 130 {
		t.Errorf("Blend(0.5) = rgb(%d,%d,%d), want ~rgb(127,127,127)", r, g, b)
	}
}

func TestStyleParse(t *testing.T) {
	tests := []struct {
		input    string
		wantBold bool
	}{
		{"bold", true},
		{"bold red", true},
		{"red", false},
		{"bold italic", true},
		{"not bold", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			s := Parse(tt.input)
			if s.IsBold() != tt.wantBold {
				t.Errorf("Parse(%q).IsBold() = %v, want %v", tt.input, s.IsBold(), tt.wantBold)
			}
		})
	}
}

func TestStyleAdd(t *testing.T) {
	base := New().WithBold(true).WithForeground(StandardColor(1))
	override := New().WithItalic(true).WithForeground(StandardColor(2))

	combined := base.Add(override)

	if !combined.IsBold() {
		t.Error("combined should be bold (inherited from base)")
	}
	if !combined.IsItalic() {
		t.Error("combined should be italic (from override)")
	}
	// Foreground should be from override (green, color 2)
	if combined.fg == nil {
		t.Error("combined should have foreground color")
	}
}

func TestStyleRender(t *testing.T) {
	tests := []struct {
		name   string
		style  Style
		text   string
		system ColorSystem
		want   string
	}{
		{
			name:   "empty style",
			style:  New(),
			text:   "hello",
			system: ColorSystemTrueColor,
			want:   "hello",
		},
		{
			name:   "bold",
			style:  New().WithBold(true),
			text:   "hello",
			system: ColorSystemTrueColor,
			want:   "\x1b[1mhello\x1b[0m",
		},
		{
			name:   "no color system",
			style:  New().WithBold(true),
			text:   "hello",
			system: ColorSystemNone,
			want:   "hello",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.style.Render(tt.text, tt.system)
			if got != tt.want {
				t.Errorf("Render() = %q, want %q", got, tt.want)
			}
		})
	}
}
