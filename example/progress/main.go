package main

import (
	"context"
	"fmt"
	"time"

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
	fmt.Println("   Done!")
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
	fmt.Println("   All tasks completed!")
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
	fmt.Println("   Download complete!")
}
