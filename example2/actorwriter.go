package example2

import (
	"context"
	"fmt"
	"io"
)

// ActorWriter writes to many WriteClosers with contexts correctly
type ActorWriter struct {
	open          func(string) (io.WriteCloser, error)
	writeRequests chan WriteRequest
	closeRequest  chan chan error
}

// NewActorWriter constructs for lazy init
func NewActorWriter(ctx context.Context,
	open func(string) (io.WriteCloser, error)) *ActorWriter {
	w := &ActorWriter{
		open:          open,
		writeRequests: make(chan WriteRequest, 100),
		closeRequest:  make(chan chan error, 1),
	}
	w.start(ctx)
	return w
}

// Start background resources
func (w ActorWriter) start(ctx context.Context) {
	go func() {
		objects := make(map[string]io.WriteCloser)
		for {
			select {
			case <-ctx.Done():
				return
			case resp := <-w.closeRequest:
				err := w.closeall(objects)
				objects = make(map[string]io.WriteCloser)
				resp <- err
			case req := <-w.writeRequests:
				f, ok := objects[req.Name]
				if !ok {
					var err error
					f, err = w.open(req.Name)
					if err != nil {
						go func(err error) {
							select {
							case <-ctx.Done():
								return
							case req.response <- WriteResponse{
								Err: err,
							}:
							}
						}(err)
						continue
					}
					objects[req.Name] = f
				}
				go func(n int, err error) {
					select {
					case <-ctx.Done():
						req.response <- WriteResponse{
							Err: ctx.Err(),
						}
					case req.response <- WriteResponse{
						Err: err,
						N:   n,
					}:
					}
				}(f.Write(req.Data))
			}
		}
	}()
}

// WriteRequest asks the actor to write on its behalf
type WriteRequest struct {
	Name     string
	Data     []byte
	response chan WriteResponse
}

// NewWriteRequest that implments Promise
func NewWriteRequest(name string, data []byte) WriteRequest {
	return WriteRequest{
		Name:     name,
		Data:     data,
		response: make(chan WriteResponse),
	}
}

// Then returns once the promise is fulfilled
func (w WriteRequest) Then(ctx context.Context) (int, error) {
	select {
	case <-ctx.Done():
		return 0, ctx.Err()
	case resp := <-w.response:
		return resp.N, resp.Err
	}
}

// Write to one of the WriteClosers
func (w ActorWriter) Write(ctx context.Context, target string,
	payload []byte) chan WriteResponse {
	req := NewWriteRequest(target, payload)
	select {
	case <-ctx.Done():
		return req.response
	case w.writeRequests <- req:
	}
	return req.response
}

// Close all the cached WriteClosers
func (w ActorWriter) Close(ctx context.Context) error {
	resp := make(chan error)
	w.closeRequest <- resp
	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-resp:
		return err
	}
}

func (w ActorWriter) closeall(objects map[string]io.WriteCloser) error {
	var mErr error
	for _, v := range objects {
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
