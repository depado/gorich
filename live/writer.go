package live

import (
	"bytes"
	"io"
	"strings"
	"sync"
)

// BlockWriter is an [io.Writer] that buffers partial lines and flushes
// complete lines to [BlockDisplay.AppendLine]. Create one per block with
// [BlockDisplay.NewWriter] and use it as the stdout/stderr sink for a child
// process. It is safe for concurrent use within a single block.
type BlockWriter struct {
	display *BlockDisplay
	idx     int
	mu      sync.Mutex
	buf     []byte
}

// NewWriter returns an io.Writer that appends completed lines to block idx.
func (d *BlockDisplay) NewWriter(idx int) *BlockWriter {
	return &BlockWriter{display: d, idx: idx}
}

// Write implements io.Writer. It is safe for concurrent use.
func (w *BlockWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.buf = append(w.buf, p...)
	for {
		i := bytes.IndexByte(w.buf, '\n')
		if i < 0 {
			break
		}
		line := string(w.buf[:i])
		w.buf = w.buf[i+1:]
		line = strings.TrimRight(line, "\r")
		w.display.AppendLine(w.idx, line)
	}
	return len(p), nil
}

// Flush emits any buffered partial line as a complete line. Call before
// Finish to ensure unterminated output is not lost.
func (w *BlockWriter) Flush() {
	w.mu.Lock()
	defer w.mu.Unlock()
	if len(w.buf) > 0 {
		w.display.AppendLine(w.idx, string(w.buf))
		w.buf = w.buf[:0]
	}
}

var _ io.Writer = (*BlockWriter)(nil)
