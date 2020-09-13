package example2

import (
	"context"
	"fmt"
	"io"
	"sync"
	"testing"
	"time"
)

func TestContextualHandlerParallel(t *testing.T) {
	defer reset()
	iterations := 100
	closeTime := time.Duration(iterations/2-35) * time.Millisecond
	timeout := make(chan struct{})
	finished := make(chan struct{})
	go func() {
		<-time.After(time.Duration(iterations)*time.Millisecond+closeTime)
		close(timeout)
	}()
	ctx, cancel := context.WithTimeout(context.Background(),
		time.Duration(iterations)*time.Millisecond+closeTime)
	defer cancel()

	go func() {
		defer func() {
			close(finished)
		}()
		writer := NewContextualWriter(func(_ string) (io.WriteCloser, error) {
			return &spyWriter{
				writeDelay: 1 * time.Millisecond,
				closeDelay: closeTime,
			}, nil
		})
		defer writer.Close()
		wg := &sync.WaitGroup{}
		for i := 0; i < iterations; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				resp := writer.Write(ctx, "test", []byte(fmt.Sprintf("%d", i)))
				select {
				case <-ctx.Done():
					return
				case <-resp:
					return
				}
			}(i)
		}
		wg.Wait()

	}()

	verify(t, iterations, timeout, finished)
}
