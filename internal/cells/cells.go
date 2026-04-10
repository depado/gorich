// Package cells provides terminal cell width calculation for Unicode text.
package cells

import "github.com/mattn/go-runewidth"

// Len returns the number of terminal cells needed to display the string.
// This accounts for double-width characters (CJK), zero-width characters,
// and other Unicode properties that affect terminal display width.
func Len(s string) int {
	return runewidth.StringWidth(s)
}

// RuneWidth returns the number of terminal cells needed to display a single rune.
func RuneWidth(r rune) int {
	return runewidth.RuneWidth(r)
}

// Truncate truncates a string to fit within the given cell width,
// appending the tail string if truncation occurs.
func Truncate(s string, width int, tail string) string {
	return runewidth.Truncate(s, width, tail)
}

// FillRight pads a string on the right to reach the given cell width.
func FillRight(s string, width int) string {
	return runewidth.FillRight(s, width)
}

// FillLeft pads a string on the left to reach the given cell width.
func FillLeft(s string, width int) string {
	return runewidth.FillLeft(s, width)
}

// Wrap wraps a string at the given cell width.
func Wrap(s string, width int) string {
	return runewidth.Wrap(s, width)
}
