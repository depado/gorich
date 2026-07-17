package main

import (
	"context"
	"flag"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/depado/gorich/console"
	"github.com/depado/gorich/live"
)

func main() {
	nBlocks := flag.Int("n", 15, "number of blocks")
	maxLines := flag.Int("l", 10, "max output lines per block")
	collapse := flag.Bool("collapse", false, "collapse finished blocks to header only")
	lastLine := flag.Bool("lastline", false, "show last output line on collapsed headers")
	flag.Parse()

	c := console.New()
	opts := []live.BlockDisplayOption{
		live.WithBlockMaxLines(*maxLines),
		live.WithBlockEllipsis(true),
		live.WithBlockSpinnerName("dots"),
		live.WithBlockPrefix("[dim]│ [/]"),
	}
	if *collapse {
		opts = append(opts, live.WithBlockCollapseOnFinish(true))
		if *lastLine {
			opts = append(opts, live.WithBlockCollapseLastLine(true))
		}
	}
	display := live.NewBlockDisplay(opts...)

	l := live.New(c, display, live.WithAutoRefresh(true), live.WithRefreshRate(15))
	ctx := context.Background()
	l.Start(ctx)

	names := []string{
		"api", "worker", "scheduler", "cache", "db-migrate", "indexer",
		"mailer", "billing", "auth", "gateway", "metrics", "cron",
	}
	messages := []func(r *rand.Rand) string{
		func(r *rand.Rand) string { return "[white]connecting to upstream[/]" },
		func(r *rand.Rand) string { return fmt.Sprintf("[dim]fetched %d records[/]", r.Intn(1000)) },
		func(r *rand.Rand) string { return fmt.Sprintf("processed batch [white]%d[/]", r.Intn(100)) },
		func(r *rand.Rand) string { return "[dim]flushing buffer[/]" },
		func(r *rand.Rand) string { return fmt.Sprintf("[red]retrying request (attempt %d)[/]", r.Intn(10)) },
		func(r *rand.Rand) string { return "warming cache" },
		func(r *rand.Rand) string { return fmt.Sprintf("[dim]committed transaction %d[/]", r.Intn(1000)) },
		func(r *rand.Rand) string { return "[white]heartbeat ok[/]" },
		func(r *rand.Rand) string { return fmt.Sprintf("queue depth [red]%d[/]", r.Intn(500)) },
		func(r *rand.Rand) string { return "[dim]rotating logs[/]" },
		func(r *rand.Rand) string {
			return fmt.Sprintf("[dim]go build -ldflags \"-X 'main.Version=v1.2.3' -X 'main.Build=%d%d%d%d' -X 'main.BuildDate=2026-07-13T01:21:58Z' -X 'main.Commit=%d%d%d%d' -X 'main.Branch=feature/very-long-branch-name-that-keeps-going' -X 'main.Builder=ci-runner-node-42' -s -w\" -gcflags=all=-l -o ./bin/some-service-with-a-long-name ./cmd/some-service-with-a-long-name[/]", r.Intn(1e9), r.Intn(1e9), r.Intn(1e9), r.Intn(1e9), r.Intn(1e9), r.Intn(1e9), r.Intn(1e9), r.Intn(1e9))
		},
	}

	var wg sync.WaitGroup
	for i := range *nBlocks {
		name := names[i%len(names)]
		if i >= len(names) {
			name = fmt.Sprintf("%s-%d", name, i/len(names)+1)
		}
		wg.Add(1)
		go func(name string, idxOffset int) {
			defer wg.Done()
			idx := display.Start(name)

			r := rand.New(rand.NewSource(time.Now().UnixNano() + int64(idxOffset)))
			ticks := 20 + r.Intn(20)
			for i := range ticks {
				time.Sleep(time.Duration(150+r.Intn(350)) * time.Millisecond)
				display.AppendLine(idx, fmt.Sprintf("[dim]%s:%-4d[/] %s", name, i+1, messages[r.Intn(len(messages))](r)))
			}

			if r.Intn(5) == 0 {
				display.Finish(idx, 1)
			} else {
				display.Finish(idx, 0)
			}
		}(name, i)
	}

	wg.Wait()
	time.Sleep(300 * time.Millisecond)
	l.Stop()
}
