package progress

import (
	"testing"
)

func TestProgressDoneIndeterminate(t *testing.T) {
	p := New(WithDisable(true))

	id := p.AddTask("Waiting...", nil)
	p.Done(id, "Connected!")

	p.mu.Lock()
	task := p.tasks[id]
	p.mu.Unlock()

	snap := task.Snapshot()
	if !snap.Finished {
		t.Error("Task should be finished after Done()")
	}
	if snap.Description != "Connected!" {
		t.Errorf("Description = %q, want %q", snap.Description, "Connected!")
	}
}

func TestProgressDoneDeterminate(t *testing.T) {
	p := New(WithDisable(true))

	total := 100.0
	id := p.AddTask("Processing", &total)
	p.Advance(id, 30)
	p.Done(id)

	p.mu.Lock()
	task := p.tasks[id]
	p.mu.Unlock()

	snap := task.Snapshot()
	if !snap.Finished {
		t.Error("Task should be finished after Done()")
	}
	if snap.Completed != 100 {
		t.Errorf("Completed = %f, want 100", snap.Completed)
	}
}

func TestProgressDoneUnknownTask(t *testing.T) {
	p := New(WithDisable(true))
	p.Done(999, "nope")
}

func TestProgressDoneMultipleTasks(t *testing.T) {
	p := New(WithDisable(true))

	id1 := p.AddTask("Task 1", nil)
	total := 50.0
	id2 := p.AddTask("Task 2", &total)

	p.Done(id1)
	p.Done(id2)

	if !p.Finished() {
		t.Error("Progress should be finished when all tasks are done")
	}
}
