package style

import (
	"strings"
)

// Attribute represents a text attribute (bold, italic, etc.).
type Attribute uint16

const (
	AttrBold Attribute = 1 << iota
	AttrDim
	AttrItalic
	AttrUnderline
	AttrBlink
	AttrBlink2
	AttrReverse
	AttrConceal
	AttrStrike
	AttrUnderline2
	AttrFrame
	AttrEncircle
	AttrOverline
)

// Style represents terminal text styling with colors and attributes.
// It uses a three-state system for attributes:
//   - Not set (will inherit from parent)
//   - Explicitly on
//   - Explicitly off
type Style struct {
	fg       *Color    // nil = not set
	bg       *Color    // nil = not set
	attrs    Attribute // which attributes are ON
	setAttrs Attribute // which attributes are explicitly SET (either on or off)
	link     string    // hyperlink URL
}

// New creates a new empty style.
func New() Style {
	return Style{}
}

// WithForeground returns a new style with the given foreground color.
func (s Style) WithForeground(c Color) Style {
	s.fg = &c
	return s
}

// WithBackground returns a new style with the given background color.
func (s Style) WithBackground(c Color) Style {
	s.bg = &c
	return s
}

// WithBold returns a new style with bold set.
func (s Style) WithBold(on bool) Style {
	return s.withAttr(AttrBold, on)
}

// WithDim returns a new style with dim set.
func (s Style) WithDim(on bool) Style {
	return s.withAttr(AttrDim, on)
}

// WithItalic returns a new style with italic set.
func (s Style) WithItalic(on bool) Style {
	return s.withAttr(AttrItalic, on)
}

// WithUnderline returns a new style with underline set.
func (s Style) WithUnderline(on bool) Style {
	return s.withAttr(AttrUnderline, on)
}

// WithBlink returns a new style with blink set.
func (s Style) WithBlink(on bool) Style {
	return s.withAttr(AttrBlink, on)
}

// WithReverse returns a new style with reverse set.
func (s Style) WithReverse(on bool) Style {
	return s.withAttr(AttrReverse, on)
}

// WithConceal returns a new style with conceal set.
func (s Style) WithConceal(on bool) Style {
	return s.withAttr(AttrConceal, on)
}

// WithStrike returns a new style with strikethrough set.
func (s Style) WithStrike(on bool) Style {
	return s.withAttr(AttrStrike, on)
}

// WithOverline returns a new style with overline set.
func (s Style) WithOverline(on bool) Style {
	return s.withAttr(AttrOverline, on)
}

// WithLink returns a new style with a hyperlink.
func (s Style) WithLink(url string) Style {
	s.link = url
	return s
}

func (s Style) withAttr(attr Attribute, on bool) Style {
	s.setAttrs |= attr
	if on {
		s.attrs |= attr
	} else {
		s.attrs &^= attr
	}
	return s
}

// Add combines two styles, with the other style taking precedence
// for any attributes that are explicitly set.
func (s Style) Add(other Style) Style {
	result := s

	// Foreground: other wins if set
	if other.fg != nil {
		result.fg = other.fg
	}

	// Background: other wins if set
	if other.bg != nil {
		result.bg = other.bg
	}

	// Attributes: combine using the three-state logic
	// Keep base attributes where other doesn't have them set
	// Override with other's attributes where other has them set
	result.attrs = (s.attrs &^ other.setAttrs) | (other.attrs & other.setAttrs)
	result.setAttrs = s.setAttrs | other.setAttrs

	// Link: other wins if set
	if other.link != "" {
		result.link = other.link
	}

	return result
}

// Foreground returns the foreground color, or nil if not set.
func (s Style) Foreground() *Color {
	return s.fg
}

// Background returns the background color, or nil if not set.
func (s Style) Background() *Color {
	return s.bg
}

// Link returns the hyperlink URL, or empty string if not set.
func (s Style) Link() string {
	return s.link
}

// IsBold returns true if bold is explicitly set to on.
func (s Style) IsBold() bool {
	return s.setAttrs&AttrBold != 0 && s.attrs&AttrBold != 0
}

// IsDim returns true if dim is explicitly set to on.
func (s Style) IsDim() bool {
	return s.setAttrs&AttrDim != 0 && s.attrs&AttrDim != 0
}

// IsItalic returns true if italic is explicitly set to on.
func (s Style) IsItalic() bool {
	return s.setAttrs&AttrItalic != 0 && s.attrs&AttrItalic != 0
}

// IsUnderline returns true if underline is explicitly set to on.
func (s Style) IsUnderline() bool {
	return s.setAttrs&AttrUnderline != 0 && s.attrs&AttrUnderline != 0
}

// IsStrike returns true if strikethrough is explicitly set to on.
func (s Style) IsStrike() bool {
	return s.setAttrs&AttrStrike != 0 && s.attrs&AttrStrike != 0
}

// IsEmpty returns true if no styling is set.
func (s Style) IsEmpty() bool {
	return s.fg == nil && s.bg == nil && s.setAttrs == 0 && s.link == ""
}

// Render wraps text with ANSI escape codes for this style.
func (s Style) Render(text string, colorSystem ColorSystem) string {
	if s.IsEmpty() || colorSystem == ColorSystemNone {
		return text
	}

	var codes []string

	// Foreground color
	if s.fg != nil && !s.fg.IsDefault() {
		c := s.fg.Downgrade(colorSystem)
		if code := c.ANSICodes(true); code != "" {
			codes = append(codes, code)
		}
	}

	// Background color
	if s.bg != nil && !s.bg.IsDefault() {
		c := s.bg.Downgrade(colorSystem)
		if code := c.ANSICodes(false); code != "" {
			codes = append(codes, code)
		}
	}

	// Text attributes
	if s.setAttrs&AttrBold != 0 && s.attrs&AttrBold != 0 {
		codes = append(codes, "1")
	}
	if s.setAttrs&AttrDim != 0 && s.attrs&AttrDim != 0 {
		codes = append(codes, "2")
	}
	if s.setAttrs&AttrItalic != 0 && s.attrs&AttrItalic != 0 {
		codes = append(codes, "3")
	}
	if s.setAttrs&AttrUnderline != 0 && s.attrs&AttrUnderline != 0 {
		codes = append(codes, "4")
	}
	if s.setAttrs&AttrBlink != 0 && s.attrs&AttrBlink != 0 {
		codes = append(codes, "5")
	}
	if s.setAttrs&AttrBlink2 != 0 && s.attrs&AttrBlink2 != 0 {
		codes = append(codes, "6")
	}
	if s.setAttrs&AttrReverse != 0 && s.attrs&AttrReverse != 0 {
		codes = append(codes, "7")
	}
	if s.setAttrs&AttrConceal != 0 && s.attrs&AttrConceal != 0 {
		codes = append(codes, "8")
	}
	if s.setAttrs&AttrStrike != 0 && s.attrs&AttrStrike != 0 {
		codes = append(codes, "9")
	}
	if s.setAttrs&AttrOverline != 0 && s.attrs&AttrOverline != 0 {
		codes = append(codes, "53")
	}

	if len(codes) == 0 && s.link == "" {
		return text
	}

	var result strings.Builder

	// Hyperlink start
	if s.link != "" {
		result.WriteString("\x1b]8;;")
		result.WriteString(s.link)
		result.WriteString("\x1b\\")
	}

	// SGR codes
	if len(codes) > 0 {
		result.WriteString("\x1b[")
		result.WriteString(strings.Join(codes, ";"))
		result.WriteString("m")
	}

	result.WriteString(text)

	// Reset SGR
	if len(codes) > 0 {
		result.WriteString("\x1b[0m")
	}

	// Hyperlink end
	if s.link != "" {
		result.WriteString("\x1b]8;;\x1b\\")
	}

	return result.String()
}

// Parse parses a style string like "bold red on blue".
func Parse(s string) Style {
	if s == "" {
		return New()
	}

	result := New()
	parts := strings.Fields(strings.ToLower(s))

	i := 0
	for i < len(parts) {
		part := parts[i]

		switch part {
		case "bold", "b":
			result = result.WithBold(true)
		case "dim":
			result = result.WithDim(true)
		case "italic", "i":
			result = result.WithItalic(true)
		case "underline", "u":
			result = result.WithUnderline(true)
		case "blink":
			result = result.WithBlink(true)
		case "reverse":
			result = result.WithReverse(true)
		case "conceal":
			result = result.WithConceal(true)
		case "strike", "s":
			result = result.WithStrike(true)
		case "overline":
			result = result.WithOverline(true)
		case "not":
			// Handle "not bold", "not italic", etc.
			if i+1 < len(parts) {
				i++
				switch parts[i] {
				case "bold", "b":
					result = result.WithBold(false)
				case "dim":
					result = result.WithDim(false)
				case "italic", "i":
					result = result.WithItalic(false)
				case "underline", "u":
					result = result.WithUnderline(false)
				case "blink":
					result = result.WithBlink(false)
				case "reverse":
					result = result.WithReverse(false)
				case "conceal":
					result = result.WithConceal(false)
				case "strike", "s":
					result = result.WithStrike(false)
				case "overline":
					result = result.WithOverline(false)
				}
			}
		case "on":
			// Background color follows
			if i+1 < len(parts) {
				i++
				if c, err := ParseColor(parts[i]); err == nil {
					result = result.WithBackground(c)
				}
			}
		case "link":
			// Link URL follows
			if i+1 < len(parts) {
				i++
				result = result.WithLink(parts[i])
			}
		default:
			// Try to parse as foreground color
			if c, err := ParseColor(part); err == nil {
				result = result.WithForeground(c)
			}
		}
		i++
	}

	return result
}

// String returns a human-readable representation of the style.
func (s Style) String() string {
	var parts []string

	if s.IsBold() {
		parts = append(parts, "bold")
	}
	if s.IsDim() {
		parts = append(parts, "dim")
	}
	if s.IsItalic() {
		parts = append(parts, "italic")
	}
	if s.IsUnderline() {
		parts = append(parts, "underline")
	}
	if s.IsStrike() {
		parts = append(parts, "strike")
	}

	if s.fg != nil {
		parts = append(parts, "fg:set")
	}
	if s.bg != nil {
		parts = append(parts, "bg:set")
	}

	if len(parts) == 0 {
		return "none"
	}
	return strings.Join(parts, " ")
}
