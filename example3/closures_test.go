package example3

import (
	"fmt"
	"testing"
)

func TestClosures(t *testing.T) {
	outside := 12
	l := func() {
		outside++
	}
	i := 0
	for ; outside < 20; i++ {
		l()
	}
	fmt.Println("outside =", outside, " i =", i)
}

func build(outside *int) func() {
	return func() {
		*outside++
	}
}

func TestClosureReturn(t *testing.T) {
	outside := 12
	l := build(&outside)
	run(l, &outside)
	fmt.Println("outside =", outside)
}

func run(l func(), outside *int) {
	i := 0
	for ; *outside < 20; i++ {
		l()
	}
	fmt.Println("i =", i)
}

func buildN(outside *int, n int) func() {
	return func() {
		*outside += n
	}
}

func TestClosure2Return(t *testing.T) {
	outside := 12
	l := buildN(&outside,2)
	run(l, &outside)
	fmt.Println("outside =", outside)
}

type CanCount interface {
	Count()
}

type VariableCounter struct {
	count func()
}

func (v VariableCounter) Count() {
	v.count()
}

func NewVariableCounter(outside *int, n int) func() CanCount {
	return func() CanCount {
		v := &VariableCounter{
			count: func() {
				*outside += n
			},
		}
		return v
	}
}

func TestClosure3Return(t *testing.T) {
	outside := 12
	counter := NewVariableCounter(&outside, 2)
	run(counter().Count, &outside)
	fmt.Println("outside =", outside)
}

type CanCountAndReport interface {
	CanCount
	Report() int
}

type VariableCounterReporter struct {
	count func()
	report func() int
}

func (v VariableCounterReporter) Count() {
	v.count()
}

func (v VariableCounterReporter) Report() int {
	return v.report()
}

func NewVariableCounterReporter(outside int, 
	n int) func() CanCountAndReport {
	return func() CanCountAndReport {
		v := &VariableCounterReporter{
			count: func() {
				outside += n
			},
			report: func() int {
				return outside
			},
		}
		return v
	}
}

func TestClosure4Return(t *testing.T) {
	outside := 12
	counter := NewVariableCounterReporter(outside, 2)
	run2(counter())
}

func run2(v CanCountAndReport) {
	i := 0
	for ; v.Report() < 20; i++ {
		v.Count()
	}
	fmt.Println("outer =", v.Report(), "i =", i)
}