package progress

import (
	"testing"
)

func TestTaskBasic(t *testing.T) {
	currentTime := 0.0
	getTime := func() float64 { return currentTime }

	total := 100.0
	task := NewTask(1, TaskConfig{
		Description: "Test task",
		Total:       &total,
		Start:       true,
	}, getTime, 30.0)

	snap := task.Snapshot()
	if snap.Description != "Test task" {
		t.Errorf("Description = %q, want %q", snap.Description, "Test task")
	}
	if !snap.Started {
		t.Error("Task should be started")
	}
	if snap.Finished {
		t.Error("Task should not be finished")
	}
	if snap.Percentage != 0 {
		t.Errorf("Percentage = %f, want 0", snap.Percentage)
	}
}

func TestTaskAdvance(t *testing.T) {
	currentTime := 0.0
	getTime := func() float64 { return currentTime }

	total := 100.0
	task := NewTask(1, TaskConfig{
		Description: "Test task",
		Total:       &total,
		Start:       true,
	}, getTime, 30.0)

	// Advance by 50
	currentTime = 1.0
	task.Advance(50)

	snap := task.Snapshot()
	if snap.Completed != 50 {
		t.Errorf("Completed = %f, want 50", snap.Completed)
	}
	if snap.Percentage != 50 {
		t.Errorf("Percentage = %f, want 50", snap.Percentage)
	}
	if snap.Finished {
		t.Error("Task should not be finished at 50%")
	}

	// Advance to completion
	currentTime = 2.0
	task.Advance(50)

	snap = task.Snapshot()
	if snap.Completed != 100 {
		t.Errorf("Completed = %f, want 100", snap.Completed)
	}
	if snap.Percentage != 100 {
		t.Errorf("Percentage = %f, want 100", snap.Percentage)
	}
	if !snap.Finished {
		t.Error("Task should be finished at 100%")
	}
}

func TestTaskSpeed(t *testing.T) {
	currentTime := 0.0
	getTime := func() float64 { return currentTime }

	total := 100.0
	task := NewTask(1, TaskConfig{
		Description: "Test task",
		Total:       &total,
		Start:       true,
	}, getTime, 30.0)

	// Advance over time to build up speed samples
	for i := 1; i <= 10; i++ {
		currentTime = float64(i)
		task.Advance(10)
	}

	snap := task.Snapshot()
	if snap.Speed == nil {
		t.Error("Speed should not be nil after multiple advances")
		return
	}

	// Speed should be approximately 10 units/second
	speed := *snap.Speed
	if speed < 9 || speed > 11 {
		t.Errorf("Speed = %f, expected ~10", speed)
	}
}

func TestTaskTimeRemaining(t *testing.T) {
	currentTime := 0.0
	getTime := func() float64 { return currentTime }

	total := 100.0
	task := NewTask(1, TaskConfig{
		Description: "Test task",
		Total:       &total,
		Start:       true,
	}, getTime, 30.0)

	// Advance at 10 units/second
	for i := 1; i <= 5; i++ {
		currentTime = float64(i)
		task.Advance(10)
	}

	// At 50% with speed of 10/s, remaining should be ~5 seconds
	snap := task.Snapshot()
	if snap.TimeRemaining == nil {
		t.Error("TimeRemaining should not be nil")
		return
	}

	remaining := *snap.TimeRemaining
	if remaining < 4 || remaining > 6 {
		t.Errorf("TimeRemaining = %f, expected ~5", remaining)
	}
}

func TestTaskReset(t *testing.T) {
	currentTime := 0.0
	getTime := func() float64 { return currentTime }

	total := 100.0
	task := NewTask(1, TaskConfig{
		Description: "Test task",
		Total:       &total,
		Start:       true,
	}, getTime, 30.0)

	currentTime = 1.0
	task.Advance(50)

	currentTime = 2.0
	task.Reset(true)

	snap := task.Snapshot()
	if snap.Completed != 0 {
		t.Errorf("Completed = %f after reset, want 0", snap.Completed)
	}
	if snap.Finished {
		t.Error("Task should not be finished after reset")
	}
}

func TestTaskIndeterminate(t *testing.T) {
	currentTime := 0.0
	getTime := func() float64 { return currentTime }

	// No total = indeterminate
	task := NewTask(1, TaskConfig{
		Description: "Test task",
		Total:       nil,
		Start:       true,
	}, getTime, 30.0)

	snap := task.Snapshot()
	if snap.Total != nil {
		t.Error("Total should be nil for indeterminate task")
	}
	if snap.Percentage != 0 {
		t.Errorf("Percentage = %f for indeterminate task, want 0", snap.Percentage)
	}
}

func TestSampleRing(t *testing.T) {
	ring := &sampleRing{}

	// Push some samples
	for i := 0; i < 5; i++ {
		ring.push(progressSample{timestamp: float64(i), completed: float64(i * 10)})
	}

	if ring.count != 5 {
		t.Errorf("count = %d, want 5", ring.count)
	}

	oldest, ok := ring.oldest()
	if !ok {
		t.Error("oldest should return true")
	}
	if oldest.timestamp != 0 {
		t.Errorf("oldest.timestamp = %f, want 0", oldest.timestamp)
	}

	newest, ok := ring.newest()
	if !ok {
		t.Error("newest should return true")
	}
	if newest.timestamp != 4 {
		t.Errorf("newest.timestamp = %f, want 4", newest.timestamp)
	}
}

func TestSampleRingOverflow(t *testing.T) {
	ring := &sampleRing{}

	// Push more than maxSamples
	for i := 0; i < maxSamples+100; i++ {
		ring.push(progressSample{timestamp: float64(i), completed: float64(i)})
	}

	if ring.count != maxSamples {
		t.Errorf("count = %d, want %d", ring.count, maxSamples)
	}

	// Oldest should be at index 100 (first 100 were overwritten)
	oldest, _ := ring.oldest()
	if oldest.timestamp != 100 {
		t.Errorf("oldest.timestamp = %f after overflow, want 100", oldest.timestamp)
	}
}
