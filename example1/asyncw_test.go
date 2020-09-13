package example1

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"testing"
	"time"
)

type canFatal interface {
	Fatal(args ...interface{})
}

type Dumper interface {
	Dump(ctx context.Context, name string) ([]byte, error)
}

func TestAsyncWriterSyncMode(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 600*time.Second)
	defer cancel()
	async, names := getWriter(ctx, t, 10, 1)
	asyncSyncMode(ctx, t, async, names, 100, true)
}

func TestAsyncWriterFullAsyncMode(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 600*time.Second)
	defer cancel()
	async, names := getWriter(ctx, t, 10, 1)
	fullAsync(ctx, t, async, names, 100, false)
}

func TestAsyncWriterBurstAsyncMode(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 600*time.Second)
	defer cancel()
	async, names := getWriter(ctx, t, 10, 1)
	burstMode(ctx, t, async, names, 100, false)
}

func getWriter(ctx context.Context, t canFatal, concurrency int,
	prepRatio int) (*AsyncWriter, map[string]chan WriteRequest) {
	async := NewAsyncWriter(NewSlowQ(prepRatio))
	async.Start(ctx)
	names := map[string]chan WriteRequest{}
	for i := 0; i < concurrency; i++ {
		names[fmt.Sprintf("%d-writer", i)] = nil
	}

	for name := range names {
		q, err := async.Writer(ctx, name)
		if err != nil {
			t.Fatal(err)
		}
		names[name] = q
	}
	return async, names
}

func asyncSyncMode(ctx context.Context, t canFatal, async *AsyncWriter, names map[string]chan WriteRequest, sendCount int, ordered bool) {
	for name, q := range names {
		for i := 0; i < sendCount; i++ {
			req := NewWriteRequest([]byte(fmt.Sprintf("%s:%d", name, i)))
			select {
			case <-ctx.Done():
				t.Fatal(ctx.Err())
			case q <- req:
			}
			req.Then(ctx, func(err error) {
				if err != nil {
					t.Fatal(err)
				}
			})
		}
	}
	if _, ok := t.(*testing.B); ok {
		return
	}
	err := checkResults(ctx, async, names, sendCount, ordered)
	if err != nil {
		t.Fatal(err)
	}
}

func burstMode(ctx context.Context, t canFatal, async *AsyncWriter, names map[string]chan WriteRequest, sendCount int, ordered bool) {
	wg := &sync.WaitGroup{}
	errors := make(chan error, 10)
	for name, q := range names {
		for j := 0; j < 10; j++ {
			for i := 0; i < sendCount; i++ {
				wg.Add(1)
				go func(name string, q chan WriteRequest, i int) {
					defer wg.Done()
					req := NewWriteRequest([]byte(fmt.Sprintf("%s:%d", name, i)))
					select {
					case <-ctx.Done():
						select {
						case errors <- ctx.Err():
						default:
							// too many forget about them
						}
						return
					case q <- req:
					}
					req.Then(ctx, func(err error) {
						if err != nil {
							select {
							case errors <- err:
							default:
								// too many forget about them
							}
						}
					})

				}(name, q, i)
			}
			time.Sleep(time.Duration(rand.Int()%1000+1000) * time.Nanosecond)
		}
	}
	wg.Wait()
	select {
	case err := <-errors:
		if err != nil {
			t.Fatal(err)
		}
	default:
	}
	if _, ok := t.(*testing.B); ok {
		return
	}
	err := checkResults(ctx, async, names, sendCount, ordered)
	if err != nil {
		t.Fatal(err)
	}
}

func checkResults(ctx context.Context, async Dumper, names map[string]chan WriteRequest, sendCount int, ordered bool) error {
	out := make(map[string][]byte)
	for name := range names {
		var err error
		out[name], err = async.Dump(ctx, name)
		if err != nil {
			return err
		}
		if out[name] == nil {
			return fmt.Errorf("expected input")
		}
		split := strings.SplitN(string(out[name]), ";", sendCount+1)
		for i, v := range split {
			if i >= sendCount {
				continue
			}
			if ordered {
				if v != fmt.Sprintf("%s:%d", name, i) {
					return fmt.Errorf("unexpected data %s for %s", v, name)
				}
			} else {
				if !strings.Contains(v, fmt.Sprintf("%s:", name)) {
					return fmt.Errorf("unexpected data %s for %s", v, name)
				}
			}
		}
	}
	return nil
}

func fullAsync(ctx context.Context, t canFatal, async *AsyncWriter, names map[string]chan WriteRequest, sendCount int, ordered bool) {
	wg := &sync.WaitGroup{}
	errors := make(chan error, 10)
	for name, q := range names {
		for i := 0; i < sendCount; i++ {
			wg.Add(1)
			go func(name string, q chan WriteRequest, i int) {
				defer wg.Done()
				req := NewWriteRequest([]byte(fmt.Sprintf("%s:%d", name, i)))
				select {
				case <-ctx.Done():
					select {
					case errors <- ctx.Err():
					default:
						// too many forget about them
					}
					return
				case q <- req:
				}
				req.Then(ctx, func(err error) {
					if err != nil {
						select {
						case errors <- err:
						default:
							// too many forget about them
						}
					}
				})

			}(name, q, i)
		}
	}
	wg.Wait()
	select {
	case err := <-errors:
		if err != nil {
			t.Fatal(err)
		}
	default:
	}
	if _, ok := t.(*testing.B); ok {
		return
	}
	err := checkResults(ctx, async, names, sendCount, ordered)
	if err != nil {
		t.Fatal(err)
	}

}
