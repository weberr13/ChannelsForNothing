package example2

import (
	"math/rand"
	"sync"
	"time"
)

func init() {
	rand.Seed(42)
}

var (
	writes int
	lock   sync.Mutex
)

func totalWrites() int {
	lock.Lock()
	defer lock.Unlock()
	return writes
}

func incWrites() {
	lock.Lock()
	writes++
	lock.Unlock()
}

func reset() {
	lock.Lock()
	writes = 0
	lock.Unlock()
}

type spyWriter struct {
	writeDelay time.Duration
	closeDelay time.Duration
}

func (s *spyWriter) Write([]byte) (int, error) {
	time.Sleep(s.writeDelay)
	incWrites()
	return 1, nil
}

func (s *spyWriter) Close() error {
	time.Sleep(s.closeDelay)
	return nil
}
