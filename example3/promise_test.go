package example3

import (
	"context"
	"sync"
	"testing"
)

func myPromiseProcessor(privateData string, threadState map[string]interface{}) func(pr Promise) {
	return func(pr Promise) {
		me := pr.(*MyPromise)
		me.data[privateData] = 1
		if _, ok := threadState["count"] ; !ok {
			threadState["count"] = []int{}
		}
		threadState["count"] = append(threadState["count"].([]int), 1)
	}
}

func myPromisePostprocess(t *testing.T, i int) func(p Promise, err error) {
	return func(p Promise, err error) {
		me := p.(*MyPromise)
		if me.data["found"].(int) != 1 || me.ident != i {
			t.Fatal("mixed up")
		}
	}
}

func TestMyPromise(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	q := make(Promiser, 100)
	st := make(chan map[string]interface{}, 100)
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
			myPromisePostprocess(t, i))
	}
	cancel()
	wg.Wait()
	state := <- st
	if len(state["count"].([]int)) != 100 {
		t.Fatal("internal state lost")
	}
}

func TestMyPromiseAsync(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	q := make(Promiser, 100)
	st := make(chan map[string]interface{}, 100)
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		myPrivateData := "found"
		state := make(map[string]interface{}) // any reference object will suffice
		q.Run(ctx, myPromiseProcessor(myPrivateData, state))
		st <- state
	}()
	workers := &sync.WaitGroup{}
	for i := 0; i < 100; i++ {
		workers.Add(1)
		go func(i int) {
			defer workers.Done()
			q.Request(MyPromiseConstructor(i)).Then(ctx, myPromisePostprocess(t,i))
		}(i)
	}
	workers.Wait()
	cancel()
	wg.Wait()
	state := <- st
	if len(state["count"].([]int)) != 100 {
		t.Fatal("internal state lost")
	}
}
