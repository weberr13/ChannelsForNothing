package example3

import (
	"context"
	"sync"
	"testing"
)

func TestMyPromise(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	q := make(Promiser, 100)
	st := make(chan map[string]interface{},1)
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		myPrivateData := "found"
		state := make(map[string]interface{})
		q.Run(ctx, myPromiseProcessor(myPrivateData, state))
		st <- state
	}()
	for i := 0; i < 100; i++ {
		q.Request(MyPromiseConstructor(i)).Then(ctx,
			myPromiseComplete(i))
	}
	cancel()
	wg.Wait()
	state := <-st
	if len(state["count"].([]int)) != 100 {
		t.Fatal("internal state lost")
	}
}

func TestMyPromiseAsync(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	q := make(Promiser, 100)
	st := make(chan map[string]interface{},1)
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		myPrivateData := "found"
		// any reference object will work
		state := make(map[string]interface{}) 
		q.Run(ctx, myPromiseProcessor(myPrivateData, state))
		st <- state
	}()
	workers := &sync.WaitGroup{}
	for i := 0; i < 100; i++ {
		workers.Add(1)
		go func(i int) {
			defer workers.Done()
			q.Request(MyPromiseConstructor(i)).Then(ctx,
				myPromiseComplete(i))
		}(i)
	}
	workers.Wait()
	cancel()
	wg.Wait()
	state := <-st
	if len(state["count"].([]int)) != 100 {
		t.Fatal("internal state lost")
	}
}
