package example2

import (
	"context"
	"fmt"
	"io"
	"sync"
	"testing"
	"time"
)

var (
	writer     *ActorWriter
	iterations = 100
	closeTime  = time.Duration(iterations/2-35) * time.Millisecond
)

func init() {
	writer = NewActorWriter(context.Background(),
		func(_ string) (io.WriteCloser, error) {
			return &spyWriter{
				writeDelay: 1 * time.Millisecond,
				closeDelay: closeTime,
			}, nil
		})
}

func TestActorHandlerParallel(t *testing.T) {
	defer reset()
	timeout := make(chan struct{})
	finished := make(chan struct{})
	ctx, cancel := context.WithTimeout(context.Background(),
		time.Duration(iterations)*time.Millisecond+closeTime)
	defer cancel()

	go func() {
		<-time.After(time.Duration(iterations)*time.Millisecond + closeTime)
		close(timeout)
	}()

	go func() {
		defer func() {
			close(finished)
		}()
		defer writer.Close(ctx)
		wg := &sync.WaitGroup{}
		for i := 0; i < iterations; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				resp := writer.Write(ctx, "test", []byte(fmt.Sprintf("%d", i)))
				wg.Add(1)
				go func() {
					defer wg.Done()
					select {
					case <-ctx.Done():
						return
					case <-resp:
						return
					}
				}()
			}(i)
		}
		wg.Wait()
	}()

	verify(t, iterations, timeout, finished)
}
