// Package style provides terminal styling with ANSI colors and text attributes.
package style

import (
	"fmt"
	"strconv"
	"strings"
)

// ColorSystem represents the color capability of a terminal.
type ColorSystem int

const (
	ColorSystemNone      ColorSystem = 0 // No color support
	ColorSystemStandard  ColorSystem = 1 // 16 colors (standard ANSI)
	ColorSystem256       ColorSystem = 2 // 256 colors
	ColorSystemTrueColor ColorSystem = 3 // 24-bit true color
)

// ColorType represents the type of color encoding.
type ColorType int

const (
	ColorTypeDefault  ColorType = iota // Default terminal color
	ColorTypeStandard                  // Standard 16 colors (0-15)
	ColorType256                       // 256 color palette (0-255)
	ColorTypeTrueColor                 // 24-bit RGB
)

// Color represents a terminal color.
type Color struct {
	typ    ColorType
	number int   // For standard/256 colors
	r, g, b uint8 // For truecolor
}

// DefaultColor returns the default terminal color.
func DefaultColor() Color {
	return Color{typ: ColorTypeDefault}
}

// StandardColor creates a color from the standard 16-color palette.
// Numbers 0-7 are normal colors, 8-15 are bright variants.
func StandardColor(n int) Color {
	if n < 0 || n > 15 {
		return DefaultColor()
	}
	return Color{typ: ColorTypeStandard, number: n}
}

// Color256 creates a color from the 256-color palette.
func Color256(n int) Color {
	if n < 0 || n > 255 {
		return DefaultColor()
	}
	return Color{typ: ColorType256, number: n}
}

// TrueColor creates a 24-bit RGB color.
func TrueColor(r, g, b uint8) Color {
	return Color{typ: ColorTypeTrueColor, r: r, g: g, b: b}
}

// RGB returns the RGB components of a color.
// For non-truecolor colors, this returns an approximation.
func (c Color) RGB() (r, g, b uint8) {
	switch c.typ {
	case ColorTypeTrueColor:
		return c.r, c.g, c.b
	case ColorType256:
		return color256ToRGB(c.number)
	case ColorTypeStandard:
		return standardColorToRGB(c.number)
	default:
		return 0, 0, 0
	}
}

// IsDefault returns true if this is the default color.
func (c Color) IsDefault() bool {
	return c.typ == ColorTypeDefault
}

// Downgrade returns a color suitable for the given color system.
func (c Color) Downgrade(system ColorSystem) Color {
	switch system {
	case ColorSystemNone:
		return DefaultColor()
	case ColorSystemStandard:
		if c.typ == ColorTypeStandard || c.typ == ColorTypeDefault {
			return c
		}
		r, g, b := c.RGB()
		return StandardColor(rgbToStandard(r, g, b))
	case ColorSystem256:
		if c.typ == ColorTypeStandard || c.typ == ColorType256 || c.typ == ColorTypeDefault {
			return c
		}
		r, g, b := c.RGB()
		return Color256(rgbTo256(r, g, b))
	default:
		return c
	}
}

// ANSICodes returns the ANSI SGR parameters for this color.
// foreground=true for foreground color, false for background.
func (c Color) ANSICodes(foreground bool) string {
	switch c.typ {
	case ColorTypeDefault:
		if foreground {
			return "39"
		}
		return "49"
	case ColorTypeStandard:
		if foreground {
			if c.number < 8 {
				return strconv.Itoa(30 + c.number)
			}
			return strconv.Itoa(90 + c.number - 8)
		}
		if c.number < 8 {
			return strconv.Itoa(40 + c.number)
		}
		return strconv.Itoa(100 + c.number - 8)
	case ColorType256:
		if foreground {
			return fmt.Sprintf("38;5;%d", c.number)
		}
		return fmt.Sprintf("48;5;%d", c.number)
	case ColorTypeTrueColor:
		if foreground {
			return fmt.Sprintf("38;2;%d;%d;%d", c.r, c.g, c.b)
		}
		return fmt.Sprintf("48;2;%d;%d;%d", c.r, c.g, c.b)
	}
	return ""
}

// Blend blends two colors together with the given factor (0.0 = c, 1.0 = other).
func (c Color) Blend(other Color, factor float64) Color {
	if factor <= 0 {
		return c
	}
	if factor >= 1 {
		return other
	}
	r1, g1, b1 := c.RGB()
	r2, g2, b2 := other.RGB()
	return TrueColor(
		blendByte(r1, r2, factor),
		blendByte(g1, g2, factor),
		blendByte(b1, b2, factor),
	)
}

func blendByte(a, b uint8, factor float64) uint8 {
	return uint8(float64(a)*(1-factor) + float64(b)*factor)
}

// Standard ANSI color names to numbers
var standardColorNames = map[string]int{
	"black":         0,
	"red":           1,
	"green":         2,
	"yellow":        3,
	"blue":          4,
	"magenta":       5,
	"cyan":          6,
	"white":         7,
	"bright_black":  8,
	"bright_red":    9,
	"bright_green":  10,
	"bright_yellow": 11,
	"bright_blue":   12,
	"bright_magenta": 13,
	"bright_cyan":   14,
	"bright_white":  15,
	"grey0":         16,
	"gray0":         16,
}

// ParseColor parses a color from a string.
// Supported formats:
//   - Named colors: "red", "bright_blue", etc.
//   - Hex: "#ff0000" or "#f00"
//   - RGB: "rgb(255,0,0)"
//   - 256-color: "color(123)"
//   - Default: "default"
func ParseColor(s string) (Color, error) {
	s = strings.TrimSpace(strings.ToLower(s))
	if s == "" || s == "default" || s == "none" {
		return DefaultColor(), nil
	}

	// Check named colors
	if n, ok := standardColorNames[s]; ok {
		return StandardColor(n), nil
	}

	// Hex color
	if strings.HasPrefix(s, "#") {
		return parseHexColor(s[1:])
	}

	// RGB function
	if strings.HasPrefix(s, "rgb(") && strings.HasSuffix(s, ")") {
		return parseRGBFunc(s[4 : len(s)-1])
	}

	// 256 color
	if strings.HasPrefix(s, "color(") && strings.HasSuffix(s, ")") {
		n, err := strconv.Atoi(strings.TrimSpace(s[6 : len(s)-1]))
		if err != nil || n < 0 || n > 255 {
			return DefaultColor(), fmt.Errorf("invalid color number: %s", s)
		}
		return Color256(n), nil
	}

	// Try as a number (256 color)
	if n, err := strconv.Atoi(s); err == nil && n >= 0 && n <= 255 {
		return Color256(n), nil
	}

	return DefaultColor(), fmt.Errorf("unknown color: %s", s)
}

func parseHexColor(s string) (Color, error) {
	if len(s) == 3 {
		// Short form: #rgb -> #rrggbb
		s = string([]byte{s[0], s[0], s[1], s[1], s[2], s[2]})
	}
	if len(s) != 6 {
		return DefaultColor(), fmt.Errorf("invalid hex color: #%s", s)
	}
	var r, g, b uint64
	var err error
	if r, err = strconv.ParseUint(s[0:2], 16, 8); err != nil {
		return DefaultColor(), err
	}
	if g, err = strconv.ParseUint(s[2:4], 16, 8); err != nil {
		return DefaultColor(), err
	}
	if b, err = strconv.ParseUint(s[4:6], 16, 8); err != nil {
		return DefaultColor(), err
	}
	return TrueColor(uint8(r), uint8(g), uint8(b)), nil
}

func parseRGBFunc(s string) (Color, error) {
	parts := strings.Split(s, ",")
	if len(parts) != 3 {
		return DefaultColor(), fmt.Errorf("invalid rgb: rgb(%s)", s)
	}
	r, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil || r < 0 || r > 255 {
		return DefaultColor(), fmt.Errorf("invalid red value: %s", parts[0])
	}
	g, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil || g < 0 || g > 255 {
		return DefaultColor(), fmt.Errorf("invalid green value: %s", parts[1])
	}
	b, err := strconv.Atoi(strings.TrimSpace(parts[2]))
	if err != nil || b < 0 || b > 255 {
		return DefaultColor(), fmt.Errorf("invalid blue value: %s", parts[2])
	}
	return TrueColor(uint8(r), uint8(g), uint8(b)), nil
}

// RGB values for standard 16 colors
var standardRGB = [][3]uint8{
	{0, 0, 0},       // black
	{128, 0, 0},     // red
	{0, 128, 0},     // green
	{128, 128, 0},   // yellow
	{0, 0, 128},     // blue
	{128, 0, 128},   // magenta
	{0, 128, 128},   // cyan
	{192, 192, 192}, // white
	{128, 128, 128}, // bright black (gray)
	{255, 0, 0},     // bright red
	{0, 255, 0},     // bright green
	{255, 255, 0},   // bright yellow
	{0, 0, 255},     // bright blue
	{255, 0, 255},   // bright magenta
	{0, 255, 255},   // bright cyan
	{255, 255, 255}, // bright white
}

func standardColorToRGB(n int) (r, g, b uint8) {
	if n < 0 || n > 15 {
		return 0, 0, 0
	}
	rgb := standardRGB[n]
	return rgb[0], rgb[1], rgb[2]
}

func color256ToRGB(n int) (r, g, b uint8) {
	if n < 0 || n > 255 {
		return 0, 0, 0
	}
	if n < 16 {
		return standardColorToRGB(n)
	}
	if n < 232 {
		// 6x6x6 color cube
		n -= 16
		r = uint8((n / 36) * 51)
		g = uint8(((n / 6) % 6) * 51)
		b = uint8((n % 6) * 51)
		return r, g, b
	}
	// Grayscale: 232-255 -> 24 shades from dark to light
	gray := uint8((n - 232) * 10 + 8)
	return gray, gray, gray
}

func rgbTo256(r, g, b uint8) int {
	// Check if it's a grayscale
	if r == g && g == b {
		if r < 8 {
			return 16 // black
		}
		if r > 248 {
			return 231 // white
		}
		return int((r-8)/10) + 232
	}
	// Map to 6x6x6 color cube
	ri := int((float64(r) / 255.0) * 5.0 + 0.5)
	gi := int((float64(g) / 255.0) * 5.0 + 0.5)
	bi := int((float64(b) / 255.0) * 5.0 + 0.5)
	return 16 + 36*ri + 6*gi + bi
}

func rgbToStandard(r, g, b uint8) int {
	// Find closest standard color by Euclidean distance
	best := 0
	bestDist := 1000000
	for i, rgb := range standardRGB {
		dr := int(r) - int(rgb[0])
		dg := int(g) - int(rgb[1])
		db := int(b) - int(rgb[2])
		dist := dr*dr + dg*dg + db*db
		if dist < bestDist {
			bestDist = dist
			best = i
		}
	}
	return best
}
