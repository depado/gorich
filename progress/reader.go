package progress

import (
	"errors"
	"io"
	"os"
	"sync/atomic"
)

// ErrNotSeekable is returned when Seek is called on a reader that doesn't support seeking.
var ErrNotSeekable = errors.New("reader does not support seeking")

// Reader wraps an io.Reader to track progress.
type Reader struct {
	reader   io.Reader
	progress *Progress
	taskID   TaskID
	read     int64 // accessed atomically
}

// NewReader creates a progress-tracking reader.
func NewReader(r io.Reader, p *Progress, taskID TaskID) *Reader {
	return &Reader{
		reader:   r,
		progress: p,
		taskID:   taskID,
	}
}

// Read implements io.Reader.
func (r *Reader) Read(p []byte) (n int, err error) {
	n, err = r.reader.Read(p)
	if n > 0 {
		atomic.AddInt64(&r.read, int64(n))
		r.progress.Advance(r.taskID, float64(n))
	}
	return n, err
}

// Close implements io.Closer if the underlying reader supports it.
func (r *Reader) Close() error {
	if closer, ok := r.reader.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}

// Seek implements io.Seeker if the underlying reader supports it.
func (r *Reader) Seek(offset int64, whence int) (int64, error) {
	seeker, ok := r.reader.(io.Seeker)
	if !ok {
		return 0, ErrNotSeekable
	}

	newPos, err := seeker.Seek(offset, whence)
	if err != nil {
		return newPos, err
	}

	// Update progress to reflect new position
	completed := float64(newPos)
	r.progress.Update(r.taskID, TaskUpdateConfig{Completed: &completed})
	atomic.StoreInt64(&r.read, newPos)

	return newPos, nil
}

// BytesRead returns the total bytes read so far.
func (r *Reader) BytesRead() int64 {
	return atomic.LoadInt64(&r.read)
}

// WrapReader wraps an io.Reader with progress tracking.
// Returns the wrapped reader and a function to stop tracking.
func (p *Progress) WrapReader(r io.Reader, total int64, description string) (*Reader, TaskID) {
	totalFloat := float64(total)
	var totalPtr *float64
	if total > 0 {
		totalPtr = &totalFloat
	}

	taskID := p.AddTask(description, totalPtr)
	return NewReader(r, p, taskID), taskID
}

// WrapFile opens a file and wraps it with progress tracking.
// Returns the wrapped reader, task ID, and any error.
func (p *Progress) WrapFile(path string, description string) (*Reader, TaskID, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, 0, err
	}

	// Get file size
	info, err := file.Stat()
	if err != nil {
		file.Close() //nolint:errcheck // cleanup on error path
		return nil, 0, err
	}

	reader, taskID := p.WrapReader(file, info.Size(), description)
	return reader, taskID, nil
}

// ReadCloser wraps an io.ReadCloser with progress tracking.
type ReadCloser struct {
	*Reader
	closer io.Closer
}

// NewReadCloser creates a progress-tracking ReadCloser.
func NewReadCloser(rc io.ReadCloser, p *Progress, taskID TaskID) *ReadCloser {
	return &ReadCloser{
		Reader: NewReader(rc, p, taskID),
		closer: rc,
	}
}

// Close closes the underlying reader.
func (rc *ReadCloser) Close() error {
	return rc.closer.Close()
}

// WrapReadCloser wraps an io.ReadCloser with progress tracking.
func (p *Progress) WrapReadCloser(rc io.ReadCloser, total int64, description string) (*ReadCloser, TaskID) {
	totalFloat := float64(total)
	var totalPtr *float64
	if total > 0 {
		totalPtr = &totalFloat
	}

	taskID := p.AddTask(description, totalPtr)
	return NewReadCloser(rc, p, taskID), taskID
}

// Writer wraps an io.Writer to track progress.
type Writer struct {
	writer   io.Writer
	progress *Progress
	taskID   TaskID
	written  int64 // accessed atomically
}

// NewWriter creates a progress-tracking writer.
func NewWriter(w io.Writer, p *Progress, taskID TaskID) *Writer {
	return &Writer{
		writer:   w,
		progress: p,
		taskID:   taskID,
	}
}

// Write implements io.Writer.
func (w *Writer) Write(p []byte) (n int, err error) {
	n, err = w.writer.Write(p)
	if n > 0 {
		atomic.AddInt64(&w.written, int64(n))
		w.progress.Advance(w.taskID, float64(n))
	}
	return n, err
}

// Close implements io.Closer if the underlying writer supports it.
func (w *Writer) Close() error {
	if closer, ok := w.writer.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}

// BytesWritten returns the total bytes written so far.
func (w *Writer) BytesWritten() int64 {
	return atomic.LoadInt64(&w.written)
}

// WrapWriter wraps an io.Writer with progress tracking.
func (p *Progress) WrapWriter(w io.Writer, total int64, description string) (*Writer, TaskID) {
	totalFloat := float64(total)
	var totalPtr *float64
	if total > 0 {
		totalPtr = &totalFloat
	}

	taskID := p.AddTask(description, totalPtr)
	return NewWriter(w, p, taskID), taskID
}

// Copy copies from src to dst while tracking progress.
// Returns the number of bytes copied and any error.
func (p *Progress) Copy(dst io.Writer, src io.Reader, total int64, description string) (int64, error) {
	reader, _ := p.WrapReader(src, total, description)
	return io.Copy(dst, reader)
}

// CopyBuffer copies from src to dst using the provided buffer while tracking progress.
func (p *Progress) CopyBuffer(dst io.Writer, src io.Reader, buf []byte, total int64, description string) (int64, error) {
	reader, _ := p.WrapReader(src, total, description)
	return io.CopyBuffer(dst, reader, buf)
}
