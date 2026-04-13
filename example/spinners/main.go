package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/depado/gorich/progress"
)

// spinnerShowcase runs a short task with the given spinner name and label.
func spinnerShowcase(name, label string) {
	p := progress.New(
		progress.WithRefreshRate(20),
		progress.WithColumns(
			progress.NewSpinnerColumn(progress.WithSpinnerName(name)),
			progress.DescriptionColumn(),
			progress.NewBarColumn(progress.WithBarWidth(30)),
			progress.NewTaskProgressColumn(false),
			progress.NewTimeElapsedColumn(),
		),
	)

	ctx := context.Background()
	p.Start(ctx)

	total := 40.0
	p.AddTask(fmt.Sprintf("[bold cyan]%-18s[/] [dim](%s)[/]", label, name), &total)

	for i := 0; i < 40; i++ {
		time.Sleep(40 * time.Millisecond)
		p.Advance(0, 1)
	}

	p.Stop()
}

// multilineDemo shows multiple concurrent tasks each with a different spinner.
func multilineDemo() {
	fmt.Println("Multiline: multiple concurrent tasks with different spinners")
	fmt.Println()

	spinnerCol := progress.NewSpinnerColumn()
	p := progress.New(
		progress.WithRefreshRate(20),
		progress.WithColumns(
			spinnerCol,
			progress.DescriptionColumn(),
			progress.NewBarColumn(progress.WithBarWidth(25)),
			progress.NewTaskProgressColumn(false),
			progress.NewMofNCompleteColumn(" / "),
			progress.NewTimeRemainingColumn(),
		),
	)

	ctx := context.Background()
	p.Start(ctx)

	type taskDef struct {
		label   string
		spinner string
		total   float64
		speed   time.Duration // delay per advance
	}

	tasks := []taskDef{
		{"[green]Compiling[/]", "dots", 80, 35 * time.Millisecond},
		{"[yellow]Linting[/]", "arc", 60, 50 * time.Millisecond},
		{"[magenta]Testing[/]", "triangle", 100, 25 * time.Millisecond},
		{"[cyan]Packaging[/]", "circleHalves", 50, 60 * time.Millisecond},
		{"[red]Deploying[/]", "growVertical", 40, 75 * time.Millisecond},
	}

	ids := make([]progress.TaskID, len(tasks))
	for i, t := range tasks {
		total := t.total
		ids[i] = p.AddTask(t.label, &total)
		spinnerCol.SetTaskSpinner(ids[i], t.spinner)
	}

	var wg sync.WaitGroup
	for i, t := range tasks {
		wg.Add(1)
		go func(id progress.TaskID, total float64, delay time.Duration) {
			defer wg.Done()
			for j := 0.0; j < total; j++ {
				time.Sleep(delay)
				p.Advance(id, 1)
			}
		}(ids[i], t.total, t.speed)
	}

	wg.Wait()
	p.Stop()
}

func main() {
	fmt.Println("GoRich Spinner Showcase")
	fmt.Println("=======================")
	fmt.Println()

	type demo struct {
		name  string
		label string
	}

	spinners := []demo{
		{"dots", "Braille dots"},
		{"dots2", "Braille dots (alt)"},
		{"line", "Classic line"},
		{"arc", "Arc"},
		{"triangle", "Triangle"},
		{"hamburger", "Hamburger"},
		{"growVertical", "Grow vertical"},
		{"growHorizontal", "Grow horizontal"},
		{"noise", "Noise"},
		{"bouncingBar", "Bouncing bar"},
		{"bouncingBall", "Bouncing ball"},
		{"pong", "Pong"},
		{"moon", "Moon phases"},
		{"earth", "Earth"},
		{"clock", "Clock"},
		{"star", "Star"},
	}

	for _, d := range spinners {
		spinnerShowcase(d.name, d.label)
	}

	fmt.Println()
	multilineDemo()
	fmt.Println()
	fmt.Println("Each task got its own spinner via spinnerCol.SetTaskSpinner().")
}
