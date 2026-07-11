package main

import (
	"context"
	"fmt"
	"time"

	"github.com/depado/gorich/console"
	"github.com/depado/gorich/progress"
)

func main() {
	fmt.Println("GoRich Progress Bar Demo")
	fmt.Println("========================")
	fmt.Println()

	// Example 1: Simple progress bar
	simpleExample()

	fmt.Println()

	// Example 2: Multiple tasks
	multipleTasksExample()

	fmt.Println()

	// Example 3: Download-style progress
	downloadExample()

	fmt.Println()

	// Example 4: Multi-section progress
	sectionExample()

	fmt.Println()

	// Example 5: Description alignment
	alignmentExample()
}

func simpleExample() {
	fmt.Println("1. Simple Progress Bar:")

	p := progress.New()
	ctx := context.Background()
	p.Start(ctx)

	total := 100.0
	task := p.AddTask("[green]Processing[/]", &total)

	for i := 0; i < 100; i++ {
		time.Sleep(20 * time.Millisecond)
		p.Advance(task, 1)
	}

	p.Stop()
}

func multipleTasksExample() {
	fmt.Println("2. Multiple Concurrent Tasks:")

	p := progress.New(progress.WithRefreshRate(30)) // 30 Hz for smoother updates
	ctx := context.Background()
	p.Start(ctx)

	total1 := 50.0
	total2 := 100.0
	total3 := 75.0

	task1 := p.AddTask("[cyan]Downloading[/]", &total1)
	task2 := p.AddTask("[yellow]Processing[/]", &total2)
	task3 := p.AddTask("[magenta]Cooking[/]", &total3)

	// Simulate concurrent work
	for i := 0; i < 100; i++ {
		time.Sleep(30 * time.Millisecond)

		if i < 50 {
			p.Advance(task1, 1)
		}
		p.Advance(task2, 1)
		if i < 75 {
			p.Advance(task3, 1)
		}
	}

	p.Stop()
}

func downloadExample() {
	fmt.Println("3. Download-Style Progress:")

	// Use download columns
	p := progress.New(
		progress.WithColumns(
			progress.NewSpinnerColumn(),
			progress.DescriptionColumn(),
			progress.NewBarColumn(progress.WithBarWidth(30)),
			progress.NewDownloadColumn(false),
			progress.NewTransferSpeedColumn(false),
			progress.NewTimeRemainingColumn(),
		),
	)

	ctx := context.Background()
	p.Start(ctx)

	// Simulate downloading 10MB file
	totalBytes := 10.0 * 1024 * 1024 // 10 MB
	task := p.AddTask("[bold blue]ubuntu-24.04.iso[/]", &totalBytes)

	// Simulate variable download speed
	downloaded := 0.0
	for downloaded < totalBytes {
		time.Sleep(50 * time.Millisecond)
		// Random-ish chunk size between 50KB and 200KB
		chunk := 100000.0 + float64(int(downloaded)%150000)
		if downloaded+chunk > totalBytes {
			chunk = totalBytes - downloaded
		}
		p.Advance(task, chunk)
		downloaded += chunk
	}

	p.Stop()
}

func sectionExample() {
	fmt.Println("4. Multi-Section Progress:")

	p := progress.New(
		progress.WithColumns(
			progress.NewSpinnerColumn(progress.WithSpinnerName("dots")),
			progress.DescriptionColumn(),
			progress.NewBarColumn(),
			progress.NewTaskProgressColumn(false),
			progress.NewSeparatorColumn("•"),
			progress.NewTimeRemainingColumn(),
		),
		progress.WithRefreshRate(30),
	)

	// A secondary section with fewer columns
	workers := p.AddSection(
		progress.WithSectionColumns(
			progress.NewSpinnerColumn(progress.WithSpinnerName("dots")),
			progress.DescriptionColumn(),
		),
	)

	ctx := context.Background()
	p.Start(ctx)

	// Default section: summary bar
	total := 88.0
	summary := p.AddTask("[bold]Syncing depado[/]", &total)

	// Simulate work
	for i := range 88 {
		time.Sleep(30 * time.Millisecond)
		p.Advance(summary, 1)

		if i == 5 {
			t := workers.AddTask("[cyan]articles - cloning...[/]", nil)
			go func() {
				time.Sleep(1500 * time.Millisecond)
				p.Done(t, "[green]articles[/]")
			}()
		}
		if i == 20 {
			t := workers.AddTask("[cyan]buoy - pulling...[/]", nil)
			go func() {
				time.Sleep(2000 * time.Millisecond)
				p.Done(t, "[green]buoy[/]")
			}()
		}
		if i == 40 {
			t := workers.AddTask("[cyan]gorich[/]", nil)
			go func() {
				time.Sleep(1200 * time.Millisecond)
				p.Done(t, "[green]gorich[/]")
			}()
		}
	}

	// Wait for workers to finish
	time.Sleep(500 * time.Millisecond)

	p.Stop()
}

func alignmentExample() {
	fmt.Println("5. Description Alignment:")

	// Left-aligned (default) auto-sizes to the widest description
	fmt.Println("   Left-aligned (default, auto-width):")
	leftAligned := progress.New(
		progress.WithColumns(
			progress.DescriptionColumn(),
			progress.NewBarColumn(progress.WithBarWidth(20)),
			progress.NewTaskProgressColumn(false),
		),
		progress.WithRefreshRate(30),
	)
	runAlignmentDemo(leftAligned)

	fmt.Println()

	// Right-aligned descriptions
	fmt.Println("   Right-aligned:")
	rightAligned := progress.New(
		progress.WithColumns(
			progress.DescriptionColumn(progress.WithJustify(console.JustifyRight)),
			progress.NewBarColumn(progress.WithBarWidth(20)),
			progress.NewTaskProgressColumn(false),
		),
		progress.WithRefreshRate(30),
	)
	runAlignmentDemo(rightAligned)
}

func runAlignmentDemo(p *progress.Progress) {
	ctx := context.Background()
	p.Start(ctx)

	total := 30.0
	t1 := p.AddTask("[cyan]Downloading[/]", &total)
	t2 := p.AddTask("[yellow]Extract[/]", &total)
	t3 := p.AddTask("[magenta]Installing packages[/]", &total)

	for range 30 {
		time.Sleep(30 * time.Millisecond)
		p.Advance(t1, 1)
		p.Advance(t2, 1)
		p.Advance(t3, 1)
	}

	p.Stop()
}
