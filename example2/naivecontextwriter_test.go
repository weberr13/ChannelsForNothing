package example2

import (
	"context"
	"fmt"
	"io"
	"sync"
	"testing"
	"time"
)

func TestContextHandler(t *testing.T) {
	defer reset()
	iterations := 100
	closeTime := time.Duration(iterations/2-35) * time.Millisecond
	timeout := make(chan struct{})
	finished := make(chan struct{})
	go func() {
		<-time.After(time.Duration(iterations*3/2) * time.Millisecond)
		close(timeout)
	}()
	ctx, cancel := context.WithTimeout(context.Background(),
		time.Duration(iterations)*time.Millisecond+closeTime)
	defer cancel()

	go func() {
		defer func() {
			close(finished)
		}()
		writer := NewCTXWriter(func(_ string) (io.WriteCloser, error) {
			return &spyWriter{
				writeDelay: 1 * time.Millisecond,
				closeDelay: closeTime,
			}, nil
		})
		defer writer.Close()
		for i := 0; i < iterations; i++ {
			writer.Write(ctx, "test", []byte(fmt.Sprintf("%d", i)))
		}

	}()

	verify(t, iterations, timeout, finished)
}

func TestParallelContextHandler(t *testing.T) {
	defer reset()
	iterations := 100
	closeTime := time.Duration(iterations/2-35) * time.Millisecond
	timeout := make(chan struct{})
	finished := make(chan struct{})
	go func() {
		<-time.After(time.Duration(iterations)*time.Millisecond + closeTime)
		close(timeout)
	}()
	ctx, cancel := context.WithTimeout(context.Background(),
		time.Duration(iterations)*time.Millisecond+closeTime)
	defer cancel()

	go func() {
		defer func() {
			close(finished)
		}()
		writer := NewCTXWriter(func(_ string) (io.WriteCloser, error) {
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
				writer.Write(ctx, "test", []byte(fmt.Sprintf("%d", i)))
			}(i)
		}
		wg.Wait()

	}()

	verify(t, iterations, timeout, finished)

}
