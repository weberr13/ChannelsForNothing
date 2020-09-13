package example1

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"
)

func getSyncWriter(concurrency int, prepRatio int) (*SyncWriter, map[string]chan WriteRequest) {
	sync := NewSyncWriter(NewSlowQ(prepRatio))
	names := map[string]chan WriteRequest{}
	for i := 0; i < concurrency; i++ {
		names[fmt.Sprintf("%d-writer", i)] = nil
	}
	return sync, names
}

func TestSyncWriterSyncMode(t *testing.T) {
	sync, names := getSyncWriter(10, 1)
	syncSync(t, sync, names, 100, true)
}

func TestSyncWriterFullAsyncMode(t *testing.T) {
	sync, names := getSyncWriter(10, 1)
	syncAsync(t, sync, names, 100, false)
}

func TestSyncWriterBurstMode(t *testing.T) {
	sync, names := getSyncWriter(10, 1)
	burstSync(t, sync, names, 100, false)
}

func syncSync(t canFatal, sync *SyncWriter, names map[string]chan WriteRequest, count int, ordered bool) {
	for i := 0; i < count; i++ {
		for name := range names {
			sync.Write(context.Background(), name, []byte(fmt.Sprintf("%s:%d", name, i)))
		}
	}
	if _, ok := t.(*testing.B); ok {
		return
	}
	err := checkResults(context.Background(), sync, names, count, ordered)
	if err != nil {
		t.Fatal(err)
	}
}

func syncAsync(t canFatal, s *SyncWriter, names map[string]chan WriteRequest, count int, ordered bool) {
	wg := &sync.WaitGroup{}
	for name := range names {
		for i := 0; i < count; i++ {
			wg.Add(1)
			go func(i int, name string) {
				defer wg.Done()
				s.Write(context.Background(), name, []byte(fmt.Sprintf("%s:%d", name, i)))
			}(i, name)
		}
	}
	wg.Wait()
	if _, ok := t.(*testing.B); ok {
		return
	}
	err := checkResults(context.Background(), s, names, count, ordered)
	if err != nil {
		t.Fatal(err)
	}
}

func burstSync(t canFatal, s *SyncWriter, names map[string]chan WriteRequest, count int, ordered bool) {
	wg := &sync.WaitGroup{}
	for name := range names {
		for j := 0; j < 10; j++ {
			for i := 0; i < count; i++ {
				wg.Add(1)
				go func(i int, name string) {
					defer wg.Done()
					s.Write(context.Background(), name, []byte(fmt.Sprintf("%s:%d", name, i)))
				}(i, name)
			}
			time.Sleep(time.Duration(rand.Int()%1000+1000) * time.Nanosecond)
		}
	}
	wg.Wait()
	if _, ok := t.(*testing.B); ok {
		return
	}
	err := checkResults(context.Background(), s, names, count, ordered)
	if err != nil {
		t.Fatal(err)
	}
}
