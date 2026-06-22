package progress

import (
	"math"
	"strings"

	"github.com/depado/gorich/console"
	"github.com/depado/gorich/segment"
	"github.com/depado/gorich/style"
)

// Bar characters
const (
	barFull      = "━"
	barHalfRight = "╸"
	barHalfLeft  = "╺"
	barEmpty     = " "
)

// ASCII fallback characters
const (
	barFullASCII  = "-"
	barEmptyASCII = " "
)

// Pulse animation constants
const (
	pulseSize      = 20   // Width of pulse in characters
	pulseSpeed     = 15.0 // Characters per second
)

// ProgressBar renders a single progress bar.
type ProgressBar struct {
	Total         *float64
	Completed     float64
	Width         *int    // nil = auto width
	Pulse         bool    // Force pulse animation
	AnimationTime float64 // Time for animation (used for pulse offset)
	ASCIIOnly     bool
	Finished      bool    // Task is finished (use finished style)

	// Style names (keys in the style map)
	BackStyle     string
	CompleteStyle string
	FinishedStyle string
	PulseStyle    string
}

// DefaultBarStyles returns the default style names for a progress bar.
func DefaultBarStyles() (back, complete, finished, pulse string) {
	return "bar.back", "bar.complete", "bar.finished", "bar.pulse"
}

// Render implements console.Renderable.
func (pb *ProgressBar) Render(c *console.Console, opts console.Options) []segment.Segment {
	width := pb.getWidth(opts)
	if width <= 0 {
		return nil
	}

	// Determine if we should pulse
	shouldPulse := pb.Pulse || pb.Total == nil

	if shouldPulse {
		return pb.renderPulse(c, opts, width)
	}

	return pb.renderDeterminate(c, opts, width)
}

func (pb *ProgressBar) getWidth(opts console.Options) int {
	if pb.Width != nil {
		return *pb.Width
	}
	// Default to 40 or max width, whichever is smaller
	w := 40
	if opts.MaxWidth > 0 && opts.MaxWidth < w {
		w = opts.MaxWidth
	}
	return w
}

func (pb *ProgressBar) renderDeterminate(c *console.Console, opts console.Options, width int) []segment.Segment {
	total := 1.0
	if pb.Total != nil && *pb.Total > 0 {
		total = *pb.Total
	}

	completed := pb.Completed
	if completed < 0 {
		completed = 0
	}
	if completed > total {
		completed = total
	}

	// Calculate fill width in half-character units (2x resolution)
	ratio := completed / total
	fillUnits := int(ratio * float64(width*2))

	fullChars := fillUnits / 2
	hasHalf := fillUnits%2 == 1
	emptyChars := width - fullChars
	if hasHalf {
		emptyChars--
	}

	// Select style based on finished state
	var barStyle style.Style
	if pb.Finished {
		barStyle = getBarStyle(pb.FinishedStyle, "bar.finished")
	} else {
		barStyle = getBarStyle(pb.CompleteStyle, "bar.complete")
	}

	// Build the bar
	var segments []segment.Segment

	// Completed portion
	if fullChars > 0 {
		char := barFull
		if pb.ASCIIOnly {
			char = barFullASCII
		}
		segments = append(segments, segment.Segment{
			Text:  strings.Repeat(char, fullChars),
			Style: &barStyle,
		})
	}

	// Half character at the junction
	if hasHalf {
		char := barHalfRight
		if pb.ASCIIOnly {
			char = barFullASCII
		}
		segments = append(segments, segment.Segment{
			Text:  char,
			Style: &barStyle,
		})
	}

	// Empty portion
	if emptyChars > 0 {
		char := barEmpty
		backStyle := getBarStyle(pb.BackStyle, "bar.back")
		segments = append(segments, segment.Segment{
			Text:  strings.Repeat(char, emptyChars),
			Style: &backStyle,
		})
	}

	return segments
}

const pulseBands = 5 // number of discrete gradient bands for pulse animation

func (pb *ProgressBar) renderPulse(c *console.Console, opts console.Options, width int) []segment.Segment {
	pulseStart := style.TrueColor(0, 0, 0)
	pulseEnd := style.TrueColor(128, 0, 255)

	offset := int(pb.AnimationTime * pulseSpeed) % (width + pulseSize)

	char := barFull
	if pb.ASCIIOnly {
		char = barFullASCII
	}

	factors := make([]float64, width)
	for i := 0; i < width; i++ {
		pulseCenter := offset - pulseSize/2
		dist := float64(i - pulseCenter)

		if dist < -float64(pulseSize)/2 || dist > float64(pulseSize)/2 {
			factors[i] = 0
		} else {
			factors[i] = (math.Cos(dist*math.Pi/float64(pulseSize)) + 1) / 2
		}
	}

	var segments []segment.Segment
	i := 0
	for i < width {
		band := int(factors[i]*float64(pulseBands-1) + 0.5)
		if band < 0 {
			band = 0
		}
		if band >= pulseBands {
			band = pulseBands - 1
		}

		j := i + 1
		for j < width {
			nextBand := int(factors[j]*float64(pulseBands-1) + 0.5)
			if nextBand < 0 {
				nextBand = 0
			}
			if nextBand >= pulseBands {
				nextBand = pulseBands - 1
			}
			if nextBand != band {
				break
			}
			j++
		}

		count := j - i
		factor := float64(band) / float64(pulseBands-1)
		color := pulseStart.Blend(pulseEnd, factor)
		s := style.New().WithForeground(color)
		segments = append(segments, segment.Segment{
			Text:  strings.Repeat(char, count),
			Style: &s,
		})

		i = j
	}

	return segments
}

// Measure implements console.Measurable.
func (pb *ProgressBar) Measure(c *console.Console, opts console.Options) console.Measurement {
	width := pb.getWidth(opts)
	return console.NewMeasurement(width, width)
}

// Helper to get a style by name, with fallback
func getBarStyle(name, fallback string) style.Style {
	// Use fallback if name is empty
	if name == "" {
		name = fallback
	}

	// For now, return hard-coded styles
	// In a full implementation, these would come from a theme
	switch name {
	case "bar.complete":
		return style.New().WithForeground(style.TrueColor(249, 38, 114)) // Magenta-ish
	case "bar.finished":
		return style.New().WithForeground(style.TrueColor(0, 255, 0)) // Green
	case "bar.back":
		return style.New().WithForeground(style.TrueColor(68, 68, 68)) // Dark gray
	case "bar.pulse":
		return style.New().WithForeground(style.TrueColor(128, 0, 255)) // Purple
	default:
		return style.New()
	}
}
