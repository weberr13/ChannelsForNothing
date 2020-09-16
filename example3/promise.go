package example3

import (
	"context"
)

// Promiser takes requests and returns promises
type Promiser chan Promise

// Request something
func (p Promiser) Request(open func() Promise) Promise {
	pr := open()
	p <- pr
	return pr
}

// Run the processor for the Promiser
func (p Promiser) Run(ctx context.Context, call func(pr Promise)) {
	for {
		select {
		case <-ctx.Done():
			return
		case pr := <-p:
			call(pr)
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

// Promise is something that does async work
type Promise interface {
	Then(ctx context.Context, call func(p Promise, err error))
	Repl() chan Promise
}

// GenericPromise struct for adding to other objects
type GenericPromise struct {
	done  bool
	reply chan Promise
}

// Init yourself called by child implementations
func (p *GenericPromise) Init() {
	p.reply = make(chan Promise)
}

// Repl channel
func (p GenericPromise) Repl() chan Promise {
	return p.reply
}

// Then call the function on itself
func (p *GenericPromise) Then(ctx context.Context, call func(p Promise, err error)) {
	select {
	case <-ctx.Done():
		call(nil, ctx.Err())
	case self := <-p.reply:
		call(self, nil)
	}
	p.done = true
}
