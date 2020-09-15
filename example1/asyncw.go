package example1

import (
	"context"
	"math/rand"
)

func init() {
	rand.Seed(42)
}

type writerRequest struct {
	name     string
	response chan chan WriteRequest
}

type dumpName struct {
	Name     string
	Response chan []byte
}

// WriteRequest includes data and a channel that promises a WriteResponse
type WriteRequest struct {
	Data     []byte
	response chan error
}

// NewWriteRequest produces a WriteRequest suitable for sending to a Writer
func NewWriteRequest(ctx context.Context, q chan WriteRequest, data []byte) WriteRequest {
	r := WriteRequest{
		Data:     data,
		response: make(chan error),
	}
	select {
	case <-ctx.Done():
		go func() {
			r.response <- ctx.Err()
		}()
		return r
	case q <- r:
	}
	return r
}

// Then calls a function based on the result of the Write
func (wr WriteRequest) Then(ctx context.Context, l func(error)) {
	select {
	case <-ctx.Done():
		l(ctx.Err())
	case err := <-wr.response:
		l(err)
	}
}

// WriteActor does writes named queues on behalf of others
type WriteActor interface {
	Start(ctx context.Context)
	Writer(ctx context.Context, name string) (chan WriteRequest, error)
	Dump(ctx context.Context, name string) ([]byte, error)
}

// AsyncWriter writes to named Queues
type AsyncWriter struct {
	newWriterRequest chan writerRequest
	dumpRequest      chan dumpName
	open             func(name string) Queue
}

// NewAsyncWriter opens named Queues and allows writers to request writes to them
func NewAsyncWriter(open func(name string) Queue) *AsyncWriter {
	return &AsyncWriter{
		newWriterRequest: make(chan writerRequest, 10),
		dumpRequest:      make(chan dumpName),
		open:             open,
	}
}

func (w AsyncWriter) dump(ctx context.Context, du dumpName, dumpers map[string]chan dumpName) {
	dumpQ, ok := dumpers[du.Name]
	if !ok {
		du.Response <- []byte{}
		return
	}
	select {
	case <-ctx.Done():
		return
	case dumpQ <- du:
	}
}

func (w AsyncWriter) receiver(ctx context.Context, inq chan WriteRequest,
	dumpQ chan dumpName, name string) {
	data := w.open(name)
	for {
		select {
		case <-ctx.Done():
			return
		case dr := <-dumpQ:
			d := data.Dump()
			c := make([]byte, len(d))
			copy(c, d)
			dr.Response <- c
		case wr := <-inq:
			data.Write(data.Prep(wr.Data))
			go func(wr WriteRequest) {
				select {
				case <-ctx.Done():
					wr.response <- ctx.Err()
					return
				case wr.response <- nil:
				}
			}(wr)
		}
	}
}

func (w AsyncWriter) actor(ctx context.Context) {
	writers := make(map[string]chan WriteRequest)
	dumpers := make(map[string]chan dumpName, 100)
	for {
		select {
		case du := <-w.dumpRequest:
			w.dump(ctx, du, dumpers)
		case <-ctx.Done():
			return
		case req := <-w.newWriterRequest:
			inq, ok := writers[req.name]
			if ok {
				req.response <- inq
				continue
			}
			dumpQ := make(chan dumpName)
			inq = make(chan WriteRequest, 1000)
			writers[req.name] = inq
			dumpers[req.name] = dumpQ
			go w.receiver(ctx, inq, dumpQ, req.name)
			req.response <- inq
		}
	}
}

// Start the background thread that manages the Queues
func (w AsyncWriter) Start(ctx context.Context) {
	go w.actor(ctx)
}

func (w AsyncWriter) Dump(ctx context.Context, name string) ([]byte, error) {
	req := dumpName{
		Name:     name,
		Response: make(chan []byte),
	}
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case w.dumpRequest <- req:
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case resp := <-req.Response:
		return resp, nil
	}
}

// Writer returns a channel for WriteRequests
func (w AsyncWriter) Writer(ctx context.Context, name string) (chan WriteRequest, error) {
	req := writerRequest{
		name:     name,
		response: make(chan chan WriteRequest),
	}
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case w.newWriterRequest <- req:
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case resp := <-req.response:
		return resp, nil
	}
}
