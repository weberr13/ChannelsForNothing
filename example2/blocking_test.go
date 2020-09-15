package example2

import (
	"fmt"
	"io"
	"testing"
	"time"
)

func TestBlockingHandler(t *testing.T) {
	defer reset()
	iterations := 100
	timeout := make(chan struct{})
	finished := make(chan struct{})
	go func() {
		<-time.After(time.Duration(iterations*3/2) * time.Millisecond)
		close(timeout)
	}()
	go func() {
		defer func() {
			close(finished)
		}()
		writer := NewRawWriter(func(_ string) (io.WriteCloser, error) {
			return &spyWriter{
				writeDelay: 1 * time.Millisecond,
				closeDelay: time.Duration(iterations/2-35) * time.Millisecond,
			}, nil
		})
		defer writer.Close()
		for i := 0; i < iterations; i++ {
			writer.Write("test", []byte(fmt.Sprintf("%d", i)))
		}

	}()

	verify(t, iterations, timeout, finished)
}

func verify(t *testing.T, iterations int, timeout chan struct{},
	finished chan struct{}) {
	select {
	case <-timeout:
		total := totalWrites()
		if total == iterations {
			t.Fatal("took too long wrote:", total, " of ", iterations,
				" but closed too late")
		} else if total > iterations {
			t.Fatal("wrote after cancel:", total, " of ", iterations)
		} else {
			t.Log("wrote:", total, " of ", iterations)
		}
	case <-finished:
		total := totalWrites()
		if total > iterations {
			t.Fatal("wrote after cancel:", total, " of ", iterations)
		} else if total < iterations {
			t.Fatal("missed some writes:", total, " of ", iterations)
		}
	}
}
