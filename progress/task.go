// Package progress provides terminal progress bars with multiple tasks and customizable columns.
package progress

import (
	"math"
	"sync"
)

// TaskID is a unique identifier for a progress task.
type TaskID int64

// progressSample is a single data point for speed estimation.
type progressSample struct {
	timestamp float64
	completed float64
}

// sampleRing is a fixed-size ring buffer for progress samples.
type sampleRing struct {
	data  [maxSamples]progressSample
	head  int // Next write position
	count int // Number of valid entries
}

const maxSamples = 1000

func (r *sampleRing) push(s progressSample) {
	r.data[r.head] = s
	r.head = (r.head + 1) % maxSamples
	if r.count < maxSamples {
		r.count++
	}
}

func (r *sampleRing) clear() {
	r.head = 0
	r.count = 0
}

// oldest returns the oldest sample, or empty if no samples.
func (r *sampleRing) oldest() (progressSample, bool) {
	if r.count == 0 {
		return progressSample{}, false
	}
	idx := (r.head - r.count + maxSamples) % maxSamples
	return r.data[idx], true
}

// newest returns the newest sample, or empty if no samples.
func (r *sampleRing) newest() (progressSample, bool) {
	if r.count == 0 {
		return progressSample{}, false
	}
	idx := (r.head - 1 + maxSamples) % maxSamples
	return r.data[idx], true
}

// iterate calls fn for each sample from oldest to newest.
func (r *sampleRing) iterate(fn func(progressSample)) {
	if r.count == 0 {
		return
	}
	start := (r.head - r.count + maxSamples) % maxSamples
	for i := 0; i < r.count; i++ {
		fn(r.data[(start+i)%maxSamples])
	}
}

// Task represents a single progress task.
type Task struct {
	mu            sync.Mutex
	id            TaskID
	description   string
	total         *float64 // nil = indeterminate
	completed     float64
	visible       bool
	fields        map[string]any
	startTime     *float64
	stopTime      *float64
	finishedTime  *float64 // Elapsed time when finished
	finishedSpeed *float64 // Speed when finished
	samples       sampleRing
	getTime       func() float64
	speedPeriod   float64 // Speed estimation window in seconds
}

// TaskConfig holds configuration for creating a task.
type TaskConfig struct {
	Description string
	Total       *float64
	Completed   float64
	Visible     bool
	Start       bool
	Fields      map[string]any
}

// NewTask creates a new task.
func NewTask(id TaskID, cfg TaskConfig, getTime func() float64, speedPeriod float64) *Task {
	t := &Task{
		id:          id,
		description: cfg.Description,
		total:       cfg.Total,
		completed:   cfg.Completed,
		visible:     cfg.Visible,
		fields:      cfg.Fields,
		getTime:     getTime,
		speedPeriod: speedPeriod,
	}
	if cfg.Start {
		now := getTime()
		t.startTime = &now
	}
	return t
}

// ID returns the task ID.
func (t *Task) ID() TaskID {
	return t.id
}

// Snapshot returns a read-only copy of the task state.
func (t *Task) Snapshot() TaskSnapshot {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := t.getTime()

	snap := TaskSnapshot{
		ID:          t.id,
		Description: t.description,
		Total:       t.total,
		Completed:   t.completed,
		Visible:     t.visible,
		Fields:      t.fields,
		CurrentTime: now,
	}

	// Elapsed time
	if t.startTime != nil {
		if t.stopTime != nil {
			elapsed := *t.stopTime - *t.startTime
			snap.Elapsed = &elapsed
		} else {
			elapsed := now - *t.startTime
			snap.Elapsed = &elapsed
		}
		snap.Started = true
	}

	// Percentage
	if t.total != nil && *t.total > 0 {
		snap.Percentage = math.Min(100.0, (t.completed / *t.total) * 100.0)
	}

	// Finished state - either finishedTime is set, or completed >= total
	if t.finishedTime != nil {
		snap.Finished = true
		snap.FinishedTime = t.finishedTime
		snap.FinishedSpeed = t.finishedSpeed
	} else if t.total != nil && t.completed >= *t.total {
		snap.Finished = true
	}

	// Speed calculation
	speed := t.speedLocked(now)
	snap.Speed = speed

	// Time remaining
	if speed != nil && *speed > 0 && t.total != nil {
		remaining := *t.total - t.completed
		if remaining > 0 {
			tr := math.Ceil(remaining / *speed)
			snap.TimeRemaining = &tr
		} else {
			zero := 0.0
			snap.TimeRemaining = &zero
		}
	}

	return snap
}

// speedLocked calculates speed from samples (must hold lock).
func (t *Task) speedLocked(now float64) *float64 {
	if t.samples.count < 2 {
		return nil
	}

	oldest, _ := t.samples.oldest()
	newest, _ := t.samples.newest()

	// Filter samples within speed period
	windowStart := now - t.speedPeriod
	totalCompleted := 0.0
	firstTimestamp := now
	lastTimestamp := 0.0

	t.samples.iterate(func(s progressSample) {
		if s.timestamp >= windowStart {
			totalCompleted += s.completed
			if s.timestamp < firstTimestamp {
				firstTimestamp = s.timestamp
			}
			if s.timestamp > lastTimestamp {
				lastTimestamp = s.timestamp
			}
		}
	})

	// Use oldest sample's timestamp as the denominator base
	if oldest.timestamp >= windowStart {
		firstTimestamp = oldest.timestamp
	}

	elapsed := lastTimestamp - firstTimestamp
	if elapsed <= 0 {
		return nil
	}

	// Speed = total steps / elapsed time
	// We subtract the oldest sample's completed value since it's the baseline
	speed := (newest.completed - oldest.completed) / elapsed
	if speed < 0 {
		speed = 0
	}

	return &speed
}

// Update updates task state.
func (t *Task) Update(cfg TaskUpdateConfig) {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := t.getTime()

	if cfg.Description != nil {
		t.description = *cfg.Description
	}
	if cfg.Total != nil {
		t.total = cfg.Total
	}
	if cfg.Completed != nil {
		t.completed = *cfg.Completed
		t.samples.push(progressSample{timestamp: now, completed: t.completed})
	}
	if cfg.Advance != nil && *cfg.Advance != 0 {
		t.completed += *cfg.Advance
		t.samples.push(progressSample{timestamp: now, completed: t.completed})
	}
	if cfg.Visible != nil {
		t.visible = *cfg.Visible
	}
	for k, v := range cfg.Fields {
		if t.fields == nil {
			t.fields = make(map[string]any)
		}
		t.fields[k] = v
	}

	// Check for finish
	t.checkFinished(now)
}

// Advance increments the completed count.
func (t *Task) Advance(amount float64) {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := t.getTime()
	t.completed += amount
	t.samples.push(progressSample{timestamp: now, completed: t.completed})
	t.checkFinished(now)
}

// checkFinished checks if the task is finished and records final state.
func (t *Task) checkFinished(now float64) {
	if t.finishedTime != nil {
		return // Already finished
	}
	if t.total == nil || t.completed < *t.total {
		return // Not finished
	}

	// Record finish state
	if t.startTime != nil {
		elapsed := now - *t.startTime
		t.finishedTime = &elapsed
	}
	t.finishedSpeed = t.speedLocked(now)
}

// Start begins timing the task.
func (t *Task) Start() {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.startTime != nil {
		return // Already started
	}
	now := t.getTime()
	t.startTime = &now
}

// Stop pauses timing the task.
func (t *Task) Stop() {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.startTime == nil || t.stopTime != nil {
		return
	}
	now := t.getTime()
	t.stopTime = &now
}

// Reset resets the task to initial state.
func (t *Task) Reset(start bool) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.completed = 0
	t.startTime = nil
	t.stopTime = nil
	t.finishedTime = nil
	t.finishedSpeed = nil
	t.samples.clear()

	if start {
		now := t.getTime()
		t.startTime = &now
	}
}

// TaskUpdateConfig holds optional fields for updating a task.
type TaskUpdateConfig struct {
	Description *string
	Total       *float64
	Completed   *float64
	Advance     *float64
	Visible     *bool
	Fields      map[string]any
}

// TaskSnapshot is a read-only copy of task state for rendering.
type TaskSnapshot struct {
	ID            TaskID
	Description   string
	Total         *float64
	Completed     float64
	Elapsed       *float64
	TimeRemaining *float64
	Percentage    float64
	Speed         *float64
	FinishedSpeed *float64
	FinishedTime  *float64
	Started       bool
	Finished      bool
	Visible       bool
	Fields        map[string]any
	CurrentTime   float64
}

// Remaining returns the remaining steps, or nil if total is unknown.
func (s TaskSnapshot) Remaining() *float64 {
	if s.Total == nil {
		return nil
	}
	r := *s.Total - s.Completed
	if r < 0 {
		r = 0
	}
	return &r
}

// GetSpeed returns the current or finished speed.
func (s TaskSnapshot) GetSpeed() *float64 {
	if s.FinishedSpeed != nil {
		return s.FinishedSpeed
	}
	return s.Speed
}
