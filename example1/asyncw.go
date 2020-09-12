package example1

import (
	"context"
	"time"
	"math/rand"
)

func init() {
	rand.Seed(42)
}

type writerRequest struct {
	name     string
	response chan chan WriteRequest
}

type internalWriteRequest struct {
	name string
	req  WriteRequest
}

type AsyncWriter struct {
	newWriterRequest chan writerRequest
	reqQueue         chan internalWriteRequest
	dumpRequest      chan DumpName
}

type WriteResponse struct {
	Err error
}

type WriteRequest struct {
	Data     []byte
	Response chan WriteResponse
}

func NewWriteRequest(data []byte) WriteRequest {
	return WriteRequest{
		Data:     data,
		Response: make(chan WriteResponse),
	}
}

func (wr WriteRequest) Join(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case resp := <-wr.Response:
		return resp.Err
	}
}

type DumpName struct {
	Name     string
	Response chan []byte
}

func NewAsyncWriter() *AsyncWriter {
	return &AsyncWriter{
		newWriterRequest: make(chan writerRequest, 10),
		reqQueue:         make(chan internalWriteRequest, 100),
		dumpRequest:      make(chan DumpName),
	}
}

func (w AsyncWriter) Start(ctx context.Context) {
	go func() {
		writers := make(map[string]chan WriteRequest)
		data := make(map[string][]byte)
		for {
			select {
			case du := <-w.dumpRequest:
				b, ok := data[du.Name]
				if !ok {
					du.Response <- []byte{}
					continue
				}
				c := make([]byte, len(b))
				copy(c, b)
				du.Response <- c
			case <-ctx.Done():
				return
			case wr := <-w.reqQueue:
				data[wr.name] = append(data[wr.name], wr.req.Data...)
				data[wr.name] = append(data[wr.name], []byte(";")...)
				time.Sleep(time.Duration(rand.Int()%1000 + 1) * time.Nanosecond)
				go func(wr internalWriteRequest) {
					select {
					case <-ctx.Done():
						wr.req.Response <- WriteResponse{
							Err: ctx.Err(),
						}
						return
					case wr.req.Response <- WriteResponse{}:
					}
				}(wr)
			case req := <-w.newWriterRequest:
				inq, ok := writers[req.name]
				if ok {
					req.response <- inq
					continue
				}
				inq = make(chan WriteRequest, 100)
				writers[req.name] = inq
				go func(inq chan WriteRequest, name string) {
					for {
						select {
						case <-ctx.Done():
							return
						case wr := <-inq:
							select {
							case <-ctx.Done():
								wr.Response <- WriteResponse{
									Err: ctx.Err(),
								}
								return
							case w.reqQueue <- internalWriteRequest{
								name: name,
								req:  wr,
							}:
							}
						}
					}
				}(inq, req.name)
				req.response <- inq
			}
		}
	}()
}

func (w AsyncWriter) Dump(ctx context.Context, name string) ([]byte, error) {
	req := DumpName{
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
