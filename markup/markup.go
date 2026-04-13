// Package markup provides Rich-style markup parsing for styled terminal text.
//
// Markup syntax:
//   - [bold]text[/bold] or [bold]text[/] - apply bold style
//   - [red]text[/] - apply color
//   - [bold red on white]text[/] - combine styles
//   - [#ff0000]text[/] - hex colors
//   - [link=https://example.com]click here[/link] - hyperlinks
//   - \\[ - escaped bracket (renders as literal [)
package markup

import (
	"regexp"
	"strings"

	"github.com/depado/gorich/internal/cells"
	"github.com/depado/gorich/segment"
	"github.com/depado/gorich/style"
)

// Tag patterns
var (
	// Matches: [tag], [tag=value], [/tag], [/]
	// Captures: (escapes)(full_tag)(tag_content)
	reTag = regexp.MustCompile(`(\\*)\[([^\]]*)\]`)
)

// Span represents a styled region of text.
type Span struct {
	Start int
	End   int
	Style style.Style
}

// Text represents parsed markup with styled spans.
type Text struct {
	Plain string // Plain text without markup
	Spans []Span // Styled regions
}

// styleEntry tracks an open style tag and where it started.
type styleEntry struct {
	style style.Style
	start int // Position in plain text where this style starts
}

// Parse parses a markup string into a Text object.
func Parse(markup string) Text {
	var plain strings.Builder
	var spans []Span
	var styleStack []styleEntry

	position := 0
	matches := reTag.FindAllStringSubmatchIndex(markup, -1)

	for _, match := range matches {
		// match indices: [full_start, full_end, escapes_start, escapes_end, content_start, content_end]
		fullStart, fullEnd := match[0], match[1]
		escapesStart, escapesEnd := match[2], match[3]
		contentStart, contentEnd := match[4], match[5]

		// Add text before this match
		if fullStart > position {
			plain.WriteString(markup[position:fullStart])
		}

		// Handle escapes
		escapes := ""
		if escapesStart >= 0 && escapesEnd > escapesStart {
			escapes = markup[escapesStart:escapesEnd]
		}

		if len(escapes) > 0 {
			// Add literal backslashes (every 2 escapes = 1 literal backslash)
			numBackslashes := len(escapes) / 2
			for i := 0; i < numBackslashes; i++ {
				plain.WriteString("\\")
			}

			// If odd number of escapes, the bracket is escaped (literal)
			if len(escapes)%2 == 1 {
				plain.WriteString("[")
				plain.WriteString(markup[contentStart:contentEnd])
				plain.WriteString("]")
				position = fullEnd
				continue
			}
		}

		// Parse the tag content
		content := markup[contentStart:contentEnd]

		if content == "/" || strings.HasPrefix(content, "/") {
			// Closing tag
			if len(styleStack) > 0 {
				// Pop style and record span
				entry := styleStack[len(styleStack)-1]
				styleStack = styleStack[:len(styleStack)-1]

				if entry.start < plain.Len() {
					spans = append(spans, Span{
						Start: entry.start,
						End:   plain.Len(),
						Style: entry.style,
					})
				}
			}
		} else {
			// Opening tag - parse as style
			s := style.Parse(content)
			styleStack = append(styleStack, styleEntry{
				style: s,
				start: plain.Len(),
			})
		}

		position = fullEnd
	}

	// Add remaining text
	if position < len(markup) {
		plain.WriteString(markup[position:])
	}

	// Close any unclosed spans
	for len(styleStack) > 0 {
		entry := styleStack[len(styleStack)-1]
		styleStack = styleStack[:len(styleStack)-1]
		if entry.start < plain.Len() {
			spans = append(spans, Span{
				Start: entry.start,
				End:   plain.Len(),
				Style: entry.style,
			})
		}
	}

	return Text{
		Plain: plain.String(),
		Spans: spans,
	}
}

// Render converts the Text to segments.
func (t Text) Render() []segment.Segment {
	if len(t.Spans) == 0 {
		return []segment.Segment{{Text: t.Plain}}
	}

	// Build a style at each character position
	// This handles overlapping spans correctly
	styles := make([]*style.Style, len(t.Plain))

	for _, span := range t.Spans {
		for i := span.Start; i < span.End && i < len(t.Plain); i++ {
			if styles[i] == nil {
				s := span.Style
				styles[i] = &s
			} else {
				combined := styles[i].Add(span.Style)
				styles[i] = &combined
			}
		}
	}

	// Group consecutive characters with the same style
	var segments []segment.Segment
	var currentText strings.Builder
	var currentStyle *style.Style

	for i, r := range t.Plain {
		charStyle := styles[i]

		if !stylesEqual(charStyle, currentStyle) {
			// Flush current segment
			if currentText.Len() > 0 {
				segments = append(segments, segment.Segment{
					Text:  currentText.String(),
					Style: currentStyle,
				})
				currentText.Reset()
			}
			currentStyle = charStyle
		}
		currentText.WriteRune(r)
	}

	// Flush final segment
	if currentText.Len() > 0 {
		segments = append(segments, segment.Segment{
			Text:  currentText.String(),
			Style: currentStyle,
		})
	}

	return segments
}

func stylesEqual(a, b *style.Style) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

// Render is a convenience function that parses and renders markup to segments.
func Render(markup string) []segment.Segment {
	return Parse(markup).Render()
}

// Escape escapes a string so it won't be interpreted as markup.
func Escape(text string) string {
	// Escape backslashes before brackets and brackets themselves
	text = strings.ReplaceAll(text, "\\", "\\\\")
	text = strings.ReplaceAll(text, "[", "\\[")
	return text
}

// Strip removes all markup tags from a string, returning plain text.
func Strip(markup string) string {
	return Parse(markup).Plain
}

// VisibleLength returns the terminal cell width of the visible text (without markup tags).
// This accounts for double-width characters (CJK), zero-width characters, etc.
func VisibleLength(markup string) int {
	return cells.Len(Strip(markup))
}
