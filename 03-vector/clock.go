package vector

import (
	"sync"
)

// A Vector clock has one slot per node.
// slot 0 = "how many events of node 0 do I know about"
// slot 1 = "how many events of node 1 do I know about", and so on.
type Vector []int

// copyOf makes a deep copy of a vector slice to prevent data races.
func copyOf(og Vector) Vector {
	copied := make(Vector, len(og))
	copy(copied, og)
	return copied
}

// Compare returns the causal relationship between two vectors:
// "before", "after", "concurrent", or "equal".
func Compare(first Vector, second Vector) string {
	firstHasSmallerSlot := false
	secondHasSmallerSlot := false

	for slot := range first {
		if first[slot] < second[slot] {
			firstHasSmallerSlot = true
		}
		if first[slot] > second[slot] {
			secondHasSmallerSlot = true
		}
	}

	if firstHasSmallerSlot && secondHasSmallerSlot {
		return "concurrent"
	}
	if firstHasSmallerSlot {
		return "before"
	}
	if secondHasSmallerSlot {
		return "after"
	}
	return "equal"
}

// Clock is one node's vector clock.
type Clock struct {
	lock   sync.Mutex
	nodeID int
	vector Vector
}

func NewClock(nodeID int, numberOfNodes int) *Clock {
	return &Clock{nodeID: nodeID, vector: make(Vector, numberOfNodes)}
}

// LocalEvent records an internal event on this node, incrementing its own slot.
func (clock *Clock) LocalEvent() Vector {
	clock.lock.Lock()
	defer clock.lock.Unlock()

	clock.vector[clock.nodeID]++
	return copyOf(clock.vector)
}

// SendMessage records a send event and returns the vector timestamp to attach to the message.
func (clock *Clock) SendMessage() Vector {
	return clock.LocalEvent()
}

// ReceiveMessage merges the incoming vector element-wise taking the max, then increments its own slot.
func (clock *Clock) ReceiveMessage(incomingVector Vector) Vector {
	clock.lock.Lock()
	defer clock.lock.Unlock()

	for slot := range clock.vector {
		if incomingVector[slot] > clock.vector[slot] {
			clock.vector[slot] = incomingVector[slot]
		}
	}

	clock.vector[clock.nodeID]++
	return copyOf(clock.vector)
}
