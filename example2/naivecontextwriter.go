package example2

import (
	"context"
	"fmt"
	"io"
	"sync"
)

// CTXWriter writes to many WriteClosers with contexts
type CTXWriter struct {
	open    func(string) (io.WriteCloser, error)
	objects map[string]io.WriteCloser
	sync.Mutex
}

// NewCTXWriter constructs for lazy init
func NewCTXWriter(open func(string) (io.WriteCloser, error)) *CTXWriter {
	return &CTXWriter{
		open:    open,
		objects: make(map[string]io.WriteCloser),
	}
}

// Write to one of the WriteClosers
func (w *CTXWriter) Write(ctx context.Context, target string,
	payload []byte) (int, error) {
	if ctx.Err() != nil {
		return 0, ctx.Err()
	}
	w.Lock()
	defer w.Unlock()
	var err error
	f, ok := w.objects[target]
	if !ok {
		f, err = w.open(target)
		if err != nil {
			return 0, err
		}
		w.objects[target] = f
	}
	return f.Write(payload)
}

// Close all the cached WriteClosers
func (w *CTXWriter) Close() error {
	w.Lock()
	defer w.Unlock()
	err := w.closeall()
	w.objects = make(map[string]io.WriteCloser)

	return err
}

func (w *CTXWriter) closeall() error {
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
