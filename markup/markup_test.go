package markup

import (
	"testing"
)

func TestParseSimple(t *testing.T) {
	text := Parse("[bold]hello[/bold]")

	if text.Plain != "hello" {
		t.Errorf("Plain = %q, want %q", text.Plain, "hello")
	}
	if len(text.Spans) != 1 {
		t.Errorf("Spans count = %d, want 1", len(text.Spans))
	}
	if text.Spans[0].Start != 0 || text.Spans[0].End != 5 {
		t.Errorf("Span = [%d:%d], want [0:5]", text.Spans[0].Start, text.Spans[0].End)
	}
}

func TestParseAutoClose(t *testing.T) {
	text := Parse("[bold]hello[/]")

	if text.Plain != "hello" {
		t.Errorf("Plain = %q, want %q", text.Plain, "hello")
	}
	if len(text.Spans) != 1 {
		t.Fatalf("Spans count = %d, want 1", len(text.Spans))
	}
	if !text.Spans[0].Style.IsBold() {
		t.Error("Style should be bold")
	}
}

func TestParseNested(t *testing.T) {
	text := Parse("[bold]hello [italic]world[/italic][/bold]")

	if text.Plain != "hello world" {
		t.Errorf("Plain = %q, want %q", text.Plain, "hello world")
	}
	if len(text.Spans) != 2 {
		t.Fatalf("Spans count = %d, want 2", len(text.Spans))
	}
}

func TestParseMultipleStyles(t *testing.T) {
	text := Parse("[bold red]hello[/]")

	if text.Plain != "hello" {
		t.Errorf("Plain = %q, want %q", text.Plain, "hello")
	}
	if len(text.Spans) != 1 {
		t.Fatalf("Spans count = %d, want 1", len(text.Spans))
	}
	if !text.Spans[0].Style.IsBold() {
		t.Error("Style should be bold")
	}
}

func TestParseNoMarkup(t *testing.T) {
	text := Parse("hello world")

	if text.Plain != "hello world" {
		t.Errorf("Plain = %q, want %q", text.Plain, "hello world")
	}
	if len(text.Spans) != 0 {
		t.Errorf("Spans count = %d, want 0", len(text.Spans))
	}
}

func TestParseEscaped(t *testing.T) {
	text := Parse("\\[not a tag]")

	if text.Plain != "[not a tag]" {
		t.Errorf("Plain = %q, want %q", text.Plain, "[not a tag]")
	}
	if len(text.Spans) != 0 {
		t.Errorf("Spans count = %d, want 0", len(text.Spans))
	}
}

func TestStrip(t *testing.T) {
	plain := Strip("[bold red]Hello[/] [italic]World[/]")

	if plain != "Hello World" {
		t.Errorf("Strip = %q, want %q", plain, "Hello World")
	}
}

func TestEscape(t *testing.T) {
	escaped := Escape("[bold]")

	if escaped != "\\[bold]" {
		t.Errorf("Escape = %q, want %q", escaped, "\\[bold]")
	}
}

func TestRender(t *testing.T) {
	segments := Render("[bold]hello[/] world")

	if len(segments) != 2 {
		t.Fatalf("Segments count = %d, want 2", len(segments))
	}

	if segments[0].Text != "hello" {
		t.Errorf("Segment 0 text = %q, want %q", segments[0].Text, "hello")
	}
	if segments[0].Style == nil || !segments[0].Style.IsBold() {
		t.Error("Segment 0 should be bold")
	}

	if segments[1].Text != " world" {
		t.Errorf("Segment 1 text = %q, want %q", segments[1].Text, " world")
	}
	if segments[1].Style != nil {
		t.Error("Segment 1 should have no style")
	}
}
