package progress

import (
	"testing"

	"github.com/depado/gorich/console"
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

func TestSpinnerColumnCleanupOnRemove(t *testing.T) {
	p := New(WithDisable(true))
	id := p.AddTask("test", nil)
	p.RemoveTask(id)
}

func TestSpinnerColumnCleanup(t *testing.T) {
	sc := NewSpinnerColumn()
	sc.Cleanup(42) // no-op if not present
	c := console.New(console.WithNoColor(true), console.WithForceTerminal(true))
	snap := TaskSnapshot{ID: 42, Total: ptrFloat(10.0), CurrentTime: 0}
	sc.Render(snap, c, c.Options()) // creates spinner for task 42
	if _, ok := sc.spinners[42]; !ok {
		t.Error("expected spinner to be created for task 42")
	}
	sc.Cleanup(42)
	if _, ok := sc.spinners[42]; ok {
		t.Error("expected spinner to be removed after Cleanup")
	}
}
