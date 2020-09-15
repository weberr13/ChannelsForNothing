package example3

import (
	"sync"
	"testing"
)

type Aggregator struct {
	finished chan *Aggregator
	stuff    []int
	total    int
	wg       *sync.WaitGroup
}

func (a *Aggregator) Count(in int) {
	a.stuff = append(a.stuff, in)
}

func (a *Aggregator) Close() {
	for _, thing := range a.stuff {
		a.total += thing
	}
	a.finished <- a
	a.wg.Done()

}

func NewAggregator(finished chan *Aggregator, wg *sync.WaitGroup) *Aggregator {
	return &Aggregator{
		finished: finished,
		wg:       wg,
	}
}

func TestMailman(t *testing.T) {
	results := make(chan *Aggregator, 100)
	wg := &sync.WaitGroup{}
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			a := NewAggregator(results, wg)
			defer a.Close()
			for j := 0; j < 100; j++ {
				a.Count(j)
			}
		}()
	}
	wg.Wait()
	for i := 0; i < 10; i++ {
		select {
		case r := <-results:
			if r.total != 4950 {
				t.Fatalf("failed to accumulate")
			}
			if len(r.stuff) != 100 {
				t.Fatal("failed to count")
			}
		default:
			t.Fatal("expected a result")
		}

	}
}
