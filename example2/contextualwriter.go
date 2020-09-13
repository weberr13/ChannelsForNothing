package example2

import (
	"context"
	"fmt"
	"io"
	"sync"
)

// ContextualWriter writes to many WriteClosers with contexts correctly
type ContextualWriter struct {
	open    func(string) (io.WriteCloser, error)
	objects map[string]io.WriteCloser
	sync.Mutex
}

// NewContextualWriter constructs for lazy init
func NewContextualWriter(open func(string) (io.WriteCloser, error)) *ContextualWriter {
	return &ContextualWriter{
		open:    open,
		objects: make(map[string]io.WriteCloser),
	}
}

// WriteResponse information
type WriteResponse struct {
	Err error
	N   int
}

// Write to one of the WriteClosers
func (w *ContextualWriter) Write(ctx context.Context, target string,
	payload []byte) chan WriteResponse {
	resp := make(chan WriteResponse)
	go w.write(ctx, resp, target, payload)
	return resp
}

func (w *ContextualWriter) write(ctx context.Context,
	promise chan WriteResponse, target string, payload []byte) {
	w.Lock()
	select {
	case <-ctx.Done():
		w.Unlock()
		return
	default:
	}
	var err error
	var n int
	f, ok := w.objects[target]
	if !ok {
		f, err = w.open(target)
		if err != nil {
			defer w.Unlock()
			select {
			case <-ctx.Done():
				return
			case promise <- WriteResponse{
				Err: err,
			}:
			}
			return
		}
		w.objects[target] = f
	}
	n, err = f.Write(payload)
	w.Unlock()
	select {
	case <-ctx.Done():
		return // the parent will be dead too
	case promise <- WriteResponse{
		Err: err,
		N:   n,
	}:
	}
}

// Close all the cached WriteClosers
func (w *ContextualWriter) Close() error {
	w.Lock()
	defer w.Unlock()
	err := w.closeall()
	w.objects = make(map[string]io.WriteCloser)

	return err
}

func (w *ContextualWriter) closeall() error {
	var mErr error
	for _, v := range w.objects {
		err := v.Close()
		if err != nil {
			if mErr == nil {
				mErr = err
				continue
			}
			mErr = fmt.Errorf("got another error %w", mErr)
		}
	}
	return mErr
}
