package vector

import (
	"fmt"
	"sync"
)

type Vector []uint64

func (v Vector) clone() Vector {
	out := make(Vector, len(v))
	copy(out, v)
	return out
}

// relation is how two vectors (and so, 2 events) relate to each other
type Relation int

const (
	Equal Relation = iota
	Before
	After
	Concurrent
)

func (r Relation) String() string {
	switch r {
	case Equal:
		return "Equal"
	case Before:
		return "Before"
	case After:
		return "After"
	case Concurrent:
		return "Concurrent"
	}
	return "Unknown"

}

// compare compares how does a relate to b
func Comprare(a, b Vector) Relation {
	if len(a) != len(b) {
		panic(fmt.Sprintf("vector length mismatchL: %d vs %d", len(a), len(b)))
	}

	aSmaller, bSmaller := false, false
	for i := range a {
		if a[i] < b[i] {
			aSmaller = true
		}
		if a[i] > b[i] {
			bSmaller = true
		}
	}
	switch {
	case aSmaller && bSmaller:
		return Concurrent
	case aSmaller:
		return Before
	case bSmaller:
		return After
	default:
		return Equal
	}
}

//clock is one node's vector clock:

type Clock struct {
	mu sync.Mutex
	id int
	v  Vector
}

// new makes a clock for node id in a  system of n nodes
func New(id int, n int) *Clock {
	return &Clock{id: id, v: make(Vector, n)}

}

// tick stamps a local event: bump only our own slot:
func (c *Clock) TickTick() Vector {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.v[c.id]++
	return c.v.clone()
}

// send stamps an outgoing message: the returned vector travels w it:
func (c *Clock) Send() Vector {
	return c.TickTick()

}

func (c *Clock) Receive(remote Vector) Vector {
	c.mu.Lock()
	defer c.mu.Unlock()
	for i := range c.v {
		if remote[i] > c.v[i] {
			c.v[i] = remote[i]
		}
	}

	c.v[c.id]++
	return c.v.clone()

}
