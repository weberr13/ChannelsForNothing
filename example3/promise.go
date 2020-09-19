package example3

import (
	"context"
)

// Complete callback for Then
type Complete func(p Promise, err error)

// Promise is something that does async work
type Promise interface {
	Then(ctx context.Context, call Complete)
	Repl() chan Promise
}

// Promiser takes requests and returns promises
type Promiser chan Promise

// Processor callback for the promise to be acted on
type Processor func(pr Promise)

// Open creates the promise
type Open func() Promise

// Request something
func (p Promiser) Request(open Open) Promise {
	pr := open()
	p <- pr
	return pr
}

// Run the processor for the Promiser
func (p Promiser) Run(ctx context.Context, work Processor) {
	for {
		select {
		case <-ctx.Done():
			return
		case pr := <-p:
			work(pr)
			go func(pr Promise) {
				select {
				case <-ctx.Done():
					return
				case pr.Repl() <- pr:
					return
				}
			}(pr)
		}
	}
}

// GenericPromise struct for adding to other objects
type GenericPromise struct {
	done  bool
	reply chan Promise
}

// Super constructs yourself. Must be called by child implementations
func (p *GenericPromise) Super() {
	p.reply = make(chan Promise)
}

// Repl channel
func (p GenericPromise) Repl() chan Promise {
	return p.reply
}

// Then call the function on itself
func (p *GenericPromise) Then(ctx context.Context, call Complete) {
	select {
	case <-ctx.Done():
		call(nil, ctx.Err())
	case self := <-p.reply:
		call(self, nil)
	}
	p.done = true
}
