package progress

import (
	"strings"
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
	snap := TaskSnapshot{ID: 42, Total: new(10.0), CurrentTime: 0}
	sc.Render(snap, c, c.Options()) // creates spinner for task 42
	if _, ok := sc.spinners[42]; !ok {
		t.Error("expected spinner to be created for task 42")
	}
	sc.Cleanup(42)
	if _, ok := sc.spinners[42]; ok {
		t.Error("expected spinner to be removed after Cleanup")
	}
}

func TestProgressSectionsString(t *testing.T) {
	p := New(
		WithConsole(console.New(console.WithNoColor(true), console.WithForceTerminal(true))),
		WithColumns(
			NewSpinnerColumn(WithSpinnerName("dots")),
			DescriptionColumn(),
			NewBarColumn(),
		),
	)

	total := 10.0
	p.AddTask("Main", &total)

	workers := p.AddSection(
		WithSectionColumns(
			NewSpinnerColumn(WithSpinnerName("dots")),
			DescriptionColumn(),
		),
		WithSectionIndent(2),
	)

	workers.AddTask("worker-a", nil)

	got := p.String()
	lines := strings.Split(strings.TrimRight(got, "\n"), "\n")

	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d: %q", len(lines), got)
	}

	if !strings.Contains(lines[0], "Main") {
		t.Errorf("line 0 should contain Main, got %q", lines[0])
	}
	if !strings.HasPrefix(lines[1], "  ") {
		t.Errorf("line 1 should be indented, got %q", lines[1])
	}
	if !strings.Contains(lines[1], "worker-a") {
		t.Errorf("line 1 should contain worker-a, got %q", lines[1])
	}
}

func TestProgressSectionsRender(t *testing.T) {
	c := console.New(console.WithNoColor(true), console.WithForceTerminal(true))
	p := New(
		WithConsole(c),
		WithColumns(
			DescriptionColumn(),
		),
	)

	total := 10.0
	p.AddTask("Main", &total)

	workers := p.AddSection(
		WithSectionColumns(
			DescriptionColumn(),
		),
		WithSectionIndent(2),
	)

	workers.AddTask("worker", nil)

	// Get renderable and render
	p.mu.Lock()
	r := p.makeRenderable()
	p.mu.Unlock()

	segments := r.Render(c, c.Options())

	var text strings.Builder
	for _, seg := range segments {
		text.WriteString(seg.Text)
	}
	got := text.String()

	lines := strings.Split(strings.TrimRight(got, "\n"), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d: %q", len(lines), got)
	}

	if !strings.HasPrefix(lines[1], "  ") {
		t.Errorf("worker line should be indented: %q", lines[1])
	}
}

func TestProgressSectionsDefaultBackwardCompat(t *testing.T) {
	p := New(
		WithColumns(
			DescriptionColumn(),
		),
	)

	total := 10.0
	id := p.AddTask("task", &total)

	// Verify task ends up in section 0
	p.mu.Lock()
	defer p.mu.Unlock()

	if len(p.sections) != 1 {
		t.Fatalf("expected 1 implicit section, got %d", len(p.sections))
	}

	if len(p.sections[0].taskOrder) != 1 {
		t.Fatalf("expected 1 task in section 0, got %d", len(p.sections[0].taskOrder))
	}
	if p.sections[0].taskOrder[0] != id {
		t.Errorf("task %d not in section 0", id)
	}
}

func TestProgressSectionsSectionAddTask(t *testing.T) {
	p := New(
		WithColumns(DescriptionColumn()),
	)

	workers := p.AddSection(
		WithSectionColumns(DescriptionColumn()),
	)

	id1 := p.AddTask("default", nil)
	id2 := workers.AddTask("worker", nil)

	p.mu.Lock()
	defer p.mu.Unlock()

	if len(p.sections) != 2 {
		t.Fatalf("expected 2 sections, got %d", len(p.sections))
	}
	if p.sections[0].taskOrder[0] != id1 {
		t.Errorf("task %d should be in section 0, got %v", id1, p.sections[0].taskOrder)
	}
	if p.sections[1].taskOrder[0] != id2 {
		t.Errorf("task %d should be in section 1, got %v", id2, p.sections[1].taskOrder)
	}
}

func TestProgressSectionsRemoveTask(t *testing.T) {
	p := New(
		WithColumns(
			NewSpinnerColumn(WithSpinnerName("dots")),
			DescriptionColumn(),
		),
	)

	workers := p.AddSection(
		WithSectionColumns(
			NewSpinnerColumn(WithSpinnerName("dots")),
			DescriptionColumn(),
		),
	)

	wID := workers.AddTask("worker", nil)

	p.mu.Lock()
	if _, ok := p.tasks[wID]; !ok {
		t.Fatal("task should exist before remove")
	}
	section := p.sectionForTask(wID)
	if section == nil || section.idx != 1 {
		t.Fatal("task should be in section 1")
	}
	p.mu.Unlock()

	p.RemoveTask(wID)

	p.mu.Lock()
	defer p.mu.Unlock()

	if _, ok := p.tasks[wID]; ok {
		t.Error("task should be removed from global map")
	}
	if len(p.sections[1].taskOrder) != 0 {
		t.Error("task should be removed from section order")
	}
}

func TestProgressSectionsEmptySectionSkipped(t *testing.T) {
	c := console.New(console.WithNoColor(true), console.WithForceTerminal(true))
	p := New(
		WithConsole(c),
		WithColumns(DescriptionColumn()),
	)

	total := 10.0
	p.AddTask("Main", &total)

	// Add a section but never add tasks to it
	workers := p.AddSection(
		WithSectionColumns(DescriptionColumn()),
		WithSectionIndent(2),
	)

	// Add another section that does have tasks
	extra := p.AddSection(
		WithSectionColumns(DescriptionColumn()),
	)
	extra.AddTask("extra", new(5.0))

	// workers is unused but declared — empty section should be skipped
	_ = workers

	p.mu.Lock()
	r := p.makeRenderable()
	p.mu.Unlock()

	segments := r.Render(c, c.Options())
	var text strings.Builder
	for _, seg := range segments {
		text.WriteString(seg.Text)
	}
	got := text.String()

	lines := strings.Split(strings.TrimRight(got, "\n"), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 rendered lines (empty section skipped), got %d: %q", len(lines), got)
	}
	if !strings.Contains(lines[0], "Main") {
		t.Errorf("line 0 should be Main: %q", lines[0])
	}
	if !strings.Contains(lines[1], "extra") {
		t.Errorf("line 1 should be extra: %q", lines[1])
	}
}

func TestProgressSectionsStringEmptySkipped(t *testing.T) {
	p := New(
		WithConsole(console.New(console.WithNoColor(true), console.WithForceTerminal(true))),
		WithColumns(DescriptionColumn()),
	)

	total := 10.0
	p.AddTask("Main", &total)

	// Empty section
	p.AddSection(WithSectionColumns(DescriptionColumn()))

	// Section with tasks
	extra := p.AddSection(WithSectionColumns(DescriptionColumn()))
	extra.AddTask("extra", nil)

	got := p.String()
	lines := strings.Split(strings.TrimRight(got, "\n"), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines (empty section skipped), got %d: %q", len(lines), got)
	}
}

func TestDescriptionColumnAutoWidth(t *testing.T) {
	p := New(
		WithConsole(console.New(console.WithNoColor(true), console.WithForceTerminal(true))),
		WithColumns(
			DescriptionColumn(),
			NewSeparatorColumn("|"),
		),
	)

	p.AddTask("Downloading", nil) // 11 chars, the widest
	p.AddTask("Processing", nil)  // 10
	p.AddTask("Cooking", nil)     // 7

	got := p.String()
	lines := strings.Split(strings.TrimRight(got, "\n"), "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d: %q", len(lines), got)
	}

	// The separator "|" must start at the same column on every line,
	// i.e. descriptions are padded to the widest (11).
	col := strings.IndexByte(lines[0], '|')
	if col < 0 {
		t.Fatalf("separator not found in %q", lines[0])
	}
	for i, line := range lines {
		if got := strings.IndexByte(line, '|'); got != col {
			t.Errorf("line %d: separator at column %d, want %d (%q)", i, got, col, line)
		}
	}

	// Widest description determines the width: "Downloading" (11) + 1 space + "|"
	if col != 12 {
		t.Errorf("separator column = %d, want 12 (11-wide desc + 1 space)", col)
	}
}

func TestDescriptionColumnAutoWidthPerSection(t *testing.T) {
	p := New(
		WithConsole(console.New(console.WithNoColor(true), console.WithForceTerminal(true))),
		WithColumns(DescriptionColumn(), NewSeparatorColumn("|")),
	)
	p.AddTask("short", nil)

	sec := p.AddSection(WithSectionColumns(DescriptionColumn(), NewSeparatorColumn("|")))
	sec.AddTask("a-much-longer-description", nil)

	got := p.String()
	lines := strings.Split(strings.TrimRight(got, "\n"), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d: %q", len(lines), got)
	}

	// Each section sizes independently: section 0 to "short" (5), section 1 to 25.
	c0 := strings.IndexByte(lines[0], '|')
	c1 := strings.IndexByte(lines[1], '|')
	if c0 == c1 {
		t.Errorf("sections should size independently, both separators at %d", c0)
	}
	if c0 != 6 {
		t.Errorf("section 0 separator = %d, want 6 (5-wide + space)", c0)
	}
}
