package example1

import (
	"math/rand"
	"time"
)

func init() {
	rand.Seed(42)
}

type SlowQ struct {
	data     []byte
	preRatio int
}

func NewSlowQ(prepRatio int) func(name string) Queue {
	return func(name string) Queue {
		return &SlowQ{
			preRatio: prepRatio,
		}
	}
}

func (s SlowQ) Prep(d []byte) []byte {
	time.Sleep(time.Duration(rand.Int()%1000+10) * time.Nanosecond)
	c := make([]byte, len(d))
	copy(c, d)
	return c
}

func (s *SlowQ) Write(d []byte) {
	time.Sleep(time.Duration(rand.Int()%(1000)+10*s.preRatio) * time.Nanosecond)
	s.data = append(s.data, d...)
	s.data = append(s.data, []byte(";")...)
}

func (s SlowQ) Dump() []byte {
	return s.data
}
