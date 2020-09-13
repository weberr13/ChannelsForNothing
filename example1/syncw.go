package example1

import (
	"context"
	"sync"
)

// Queue that writes bytes someplace
type Queue interface {
	// Prep is a pure function that performs transformations on input
	Prep(d []byte) []byte
	// Write is not threadsafe and interacts with the greater universe
	Write(d []byte)
	// Dump is not threadsafe and returns all the data written
	Dump() []byte
}

// Writer to an embedded Queue object
type Writer interface {
	Write(ctx context.Context, name string, input []byte) error
	Dump(ctx context.Context, name string) ([]byte, error)
}

// SyncWriter writes using a standard mutex for threadsafety
type SyncWriter struct {
	data map[string]Queue
	open func(name string) Queue
	sync.Mutex
}

// NewSyncWriter writes to named Queues knowing they are not threadsafe themselves
func NewSyncWriter(open func(name string) Queue) *SyncWriter {
	return &SyncWriter{
		data: make(map[string]Queue),
		open: open,
	}
}

// Write to the named queue
func (sw *SyncWriter) Write(ctx context.Context, name string, input []byte) error {
	sw.Lock()
	defer sw.Unlock()
	if ctx.Err() != nil {
		return ctx.Err()
	}
	q, ok := sw.data[name]
	if !ok {
		sw.data[name] = sw.open(name)
		q = sw.data[name]
	}
	d := q.Prep(input)
	if ctx.Err() != nil {
		return ctx.Err()
	}
	// Simulate slow writes
	sw.data[name].Write(d)
	return nil
}

// Dump the data in the queue
func (sw *SyncWriter) Dump(ctx context.Context, name string) ([]byte, error) {
	sw.Lock()
	defer sw.Unlock()
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	q, ok := sw.data[name]
	if !ok {
		return []byte{}, nil
	}
	b := q.Dump()
	c := make([]byte, len(b))
	copy(c, b)
	return c, nil
}
