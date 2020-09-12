package example1

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"
)


type canFatal interface {
	Fatal(args ...interface{})
}

func TestAsyncWriterSyncMode(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 600*time.Second)
	defer cancel()
	async,names := getWriter(ctx, t)
	syncMode(ctx, t, async, names, 100, true, false)
}

func BenchmarkSyncMode(b *testing.B) {
	ctx, cancel := context.WithTimeout(context.Background(), 600*time.Second)
	async,names := getWriter(ctx, b)
	b.ResetTimer()
	syncMode(ctx, b, async, names, b.N, true, true)
	cancel()
}

func TestAsyncWriterFullAsyncMode(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 600*time.Second)
	defer cancel()
	async,names := getWriter(ctx, t)
	fullAsync(ctx, t, async, names, 100, false, false)
}

func BenchmarkFullAsync(b *testing.B) {
	ctx, cancel := context.WithTimeout(context.Background(), 600*time.Second)
	async,names := getWriter(ctx, b)
	b.ResetTimer()
	fullAsync(ctx, b, async, names, b.N, false, true)
	cancel()
}

func getWriter(ctx context.Context, t canFatal) (*AsyncWriter, map[string]chan WriteRequest) {
	async := NewAsyncWriter()
	async.Start(ctx)
	names := map[string]chan WriteRequest{"foo": nil, "bar": nil}

	for name := range names {
		q, err := async.Writer(ctx, name)
		if err != nil {
			t.Fatal(err)
		}
		names[name] = q
	}
	return async, names
}

func syncMode(ctx context.Context, t canFatal, async *AsyncWriter, names map[string]chan WriteRequest, sendCount int, ordered bool, bench bool) {
	for i := 0; i < sendCount; i++ {
		for name, q := range names {
			req := NewWriteRequest([]byte(fmt.Sprintf("%s:%d", name, i)))
			select {
			case <-ctx.Done():
				t.Fatal(ctx.Err())
			case q <- req:
			}
			err := req.Join(ctx)
			if err != nil {
				t.Fatal(err)
			}
		}
	}
	if bench {
		return
	}
	err := checkResults(ctx, async, names, sendCount, ordered)
	if err != nil {
		t.Fatal(err)
	}
}

func checkResults(ctx context.Context, async *AsyncWriter, names map[string]chan WriteRequest, sendCount int, ordered bool) error {
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

func fullAsync(ctx context.Context, t canFatal, async *AsyncWriter, names map[string]chan WriteRequest, sendCount int, ordered bool, bench bool) {
	wg := &sync.WaitGroup{}
	errors := make(chan error, 10)
	for i := 0; i < sendCount; i++ {
		for name, q := range names {
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
				err := req.Join(ctx)
				if err != nil {
					select {
					case errors <- err:
					default:
						// too many forget about them
					}
				}
			}(name, q, i)
		}
	}
	if bench {
		return
	}
	wg.Wait()
	select {
	case err := <-errors:
		if err != nil {
			t.Fatal(err)
		}
	default:
	}
	err := checkResults(ctx, async, names, sendCount, ordered)
	if err != nil {
		t.Fatal(err)
	}

}

