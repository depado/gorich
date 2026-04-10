package spinner

import (
	"github.com/depado/gorich/console"
	"github.com/depado/gorich/segment"
	"github.com/depado/gorich/style"
)

// Spinner renders an animated spinner.
type Spinner struct {
	Name      string
	Style     *style.Style
	Speed     float64 // Multiplier for animation speed (1.0 = normal)
	startTime float64
}

// New creates a new Spinner with the given name.
func New(name string) *Spinner {
	if name == "" {
		name = Default
	}
	return &Spinner{
		Name:  name,
		Speed: 1.0,
	}
}

// WithStyle sets the spinner style.
func (s *Spinner) WithStyle(st style.Style) *Spinner {
	s.Style = &st
	return s
}

// WithSpeed sets the animation speed multiplier.
func (s *Spinner) WithSpeed(speed float64) *Spinner {
	s.Speed = speed
	return s
}

// Render renders the spinner at the given time.
func (s *Spinner) Render(currentTime float64) []segment.Segment {
	def := Get(s.Name)

	if len(def.Frames) == 0 {
		return nil
	}

	// Calculate which frame to show
	elapsed := currentTime - s.startTime
	if s.startTime == 0 {
		s.startTime = currentTime
		elapsed = 0
	}

	intervalSec := def.Interval.Seconds()
	if s.Speed > 0 {
		intervalSec /= s.Speed
	}

	frameIndex := int(elapsed/intervalSec) % len(def.Frames)
	frame := def.Frames[frameIndex]

	return []segment.Segment{segment.NewText(frame, s.Style)}
}

// RenderConsole implements console.Renderable.
func (s *Spinner) RenderConsole(c *console.Console, opts console.Options, currentTime float64) []segment.Segment {
	return s.Render(currentTime)
}

// Reset resets the spinner animation to the beginning.
func (s *Spinner) Reset() {
	s.startTime = 0
}

// FrameCount returns the number of frames in this spinner.
func (s *Spinner) FrameCount() int {
	return len(Get(s.Name).Frames)
}

// Interval returns the interval between frames.
func (s *Spinner) Interval() float64 {
	return Get(s.Name).Interval.Seconds()
}
