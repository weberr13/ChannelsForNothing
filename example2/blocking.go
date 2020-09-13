package example2

import (
	"fmt"
	"io"
	"sync"
)

// RawWriter writes to many WriteClosers
type RawWriter struct {
	open    func(string) (io.WriteCloser, error)
	objects map[string]io.WriteCloser
	sync.Mutex
}

// NewRawWriter constructs for lazy init
func NewRawWriter(open func(string) (io.WriteCloser, error)) *RawWriter {
	return &RawWriter{
		open:    open,
		objects: make(map[string]io.WriteCloser),
	}
}

// Write to one of the WriteClosers
func (w *RawWriter) Write(target string, payload []byte) (int, error) {
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
func (w *RawWriter) Close() error {
	w.Lock()
	defer w.Unlock()
	err := w.closeall()
	w.objects = make(map[string]io.WriteCloser)

	return err
}

func (w *RawWriter) closeall() error {
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
