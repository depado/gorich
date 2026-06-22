package table

import (
	"strings"
	"testing"

	"github.com/depado/gorich/console"
	"github.com/depado/gorich/segment"
	"github.com/depado/gorich/style"
	"github.com/depado/gorich/table/box"
)

func TestTableBasicRender(t *testing.T) {
	tbl := NewTable("Name", "Age")
	tbl.AddRow("Alice", "30")
	tbl.AddRow("Bob", "25")

	c := console.New(console.WithWidth(80), console.WithNoColor(true), console.WithForceTerminal(true))

	result := tbl.Render(c, c.Options())
	output := segmentsToString(result, c.ColorSystem())
	if !strings.Contains(output, "Name") || !strings.Contains(output, "Alice") {
		t.Errorf("basic table render missing content:\n%s", output)
	}
}

func TestTableNoColumns(t *testing.T) {
	tbl := NewTable()
	c := console.New(console.WithWidth(80), console.WithNoColor(true))
	result := tbl.Render(c, c.Options())
	output := segmentsToString(result, c.ColorSystem())
	if output != "\n" {
		t.Errorf("empty table should render just newline, got %q", output)
	}
}

func TestTableWithTitle(t *testing.T) {
	tbl := NewTableWithOptions([]string{"Col1", "Col2"}, WithTitle("My Table"))
	tbl.AddRow("a", "b")

	c := console.New(console.WithWidth(80), console.WithNoColor(true), console.WithForceTerminal(true))
	result := tbl.Render(c, c.Options())
	output := segmentsToString(result, c.ColorSystem())
	if !strings.Contains(output, "My Table") {
		t.Errorf("table with title missing title:\n%s", output)
	}
}

func TestTableMeasure(t *testing.T) {
	tbl := NewTable("A", "B", "C")
	tbl.AddRow("x", "y", "z")

	c := console.New(console.WithWidth(80), console.WithNoColor(true))
	meas := tbl.Measure(c, c.Options())
	if meas.Minimum <= 0 {
		t.Errorf("table measure should be positive, got %d", meas.Minimum)
	}
}

func TestTableBoxSubstitution(t *testing.T) {
	tbl := NewTableWithOptions([]string{"Col"}, WithBox(box.HEAVY_HEAD))
	tbl.AddRow("data")

	c := console.New(console.WithWidth(80), console.WithNoColor(true), console.WithForceTerminal(true))
	// With showHeader=false, HEAVY_HEAD should substitute to SQUARE
	tbl.showHeader = false
	result := tbl.Render(c, c.Options())
	output := segmentsToString(result, c.ColorSystem())
	if !strings.Contains(output, "┌") || !strings.Contains(output, "└") {
		t.Errorf("showHeader=false should substitute to SQUARE box:\n%s", output)
	}
}

func TestTableAddSection(t *testing.T) {
	tbl := NewTable("Col")
	tbl.AddRow("row1")
	tbl.AddSection()
	tbl.AddRow("row2")

	c := console.New(console.WithWidth(80), console.WithNoColor(true), console.WithForceTerminal(true))
	result := tbl.Render(c, c.Options())
	output := segmentsToString(result, c.ColorSystem())
	if !strings.Contains(output, "row1") || !strings.Contains(output, "row2") {
		t.Errorf("table with sections missing data:\n%s", output)
	}
}

func TestTableExpand(t *testing.T) {
	tbl := NewTableWithOptions([]string{"Col"}, WithExpand())
	tbl.AddRow("data")

	c := console.New(console.WithWidth(100), console.WithNoColor(true), console.WithForceTerminal(true))
	result := tbl.Render(c, c.Options())
	output := segmentsToString(result, c.ColorSystem())
	if output == "" {
		t.Error("expand table produced no output")
	}
}

func TestTableFixedWidth(t *testing.T) {
	tbl := NewTableWithOptions([]string{"Col1", "Col2"}, WithWidth(40))
	tbl.AddRow("a", "b")

	c := console.New(console.WithWidth(100), console.WithNoColor(true), console.WithForceTerminal(true))
	meas := tbl.Measure(c, c.Options())
	if meas.Maximum > 40 {
		t.Errorf("fixed width table measure should be <= 40, got %d", meas.Maximum)
	}
}

func TestTableAddRowNonStringValues(t *testing.T) {
	c := console.New(console.WithWidth(80), console.WithNoColor(true), console.WithForceTerminal(true))
	tbl := NewTable("Col1", "Col2")
	tbl.AddRow(42, 3.14)

	result := tbl.Render(c, c.Options())
	var output strings.Builder
	for _, seg := range result {
		output.WriteString(seg.Text)
	}
	out := output.String()

	if !strings.Contains(out, "42") {
		t.Errorf("expected rendered output to contain '42', got: %s", out)
	}
	if !strings.Contains(out, "3.14") {
		t.Errorf("expected rendered output to contain '3.14', got: %s", out)
	}
}

func TestTableAddStyledRowNonStringValues(t *testing.T) {
	c := console.New(console.WithWidth(80), console.WithNoColor(true), console.WithForceTerminal(true))
	tbl := NewTable("Col1")
	tbl.AddStyledRow([]interface{}{true, 100}, nil, false)

	result := tbl.Render(c, c.Options())
	var output strings.Builder
	for _, seg := range result {
		output.WriteString(seg.Text)
	}
	out := output.String()

	if !strings.Contains(out, "true") {
		t.Errorf("expected rendered output to contain 'true', got: %s", out)
	}
	if !strings.Contains(out, "100") {
		t.Errorf("expected rendered output to contain '100', got: %s", out)
	}
}

func segmentsToString(segs []segment.Segment, cs style.ColorSystem) string {
	var sb strings.Builder
	for _, seg := range segs {
		sb.WriteString(seg.Render(cs))
	}
	return sb.String()
}
