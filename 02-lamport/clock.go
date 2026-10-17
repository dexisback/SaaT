package lamport

import (
	"sync"
)

// TimeStamp is a Lamport timestamp: the logical counter, and the ID of the node that stamped it.
type TimeStamp struct {
	Counter int
	NodeID  int
}

// IsBefore says whether one TimeStamp is ordered before another TimeStamp.
// Counters are compared first; ties are broken deterministically by NodeID.
func (first TimeStamp) IsBefore(second TimeStamp) bool {
	if first.Counter != second.Counter {
		return first.Counter < second.Counter
	}
	return first.NodeID < second.NodeID
}

// isBefore maintains backwards compatibility
func (first TimeStamp) isBefore(second TimeStamp) bool {
	return first.IsBefore(second)
}

// Clock is one node's Lamport logical clock.
type Clock struct {
	lock    sync.Mutex
	nodeID  int
	Counter int
}

func NewClock(nodeID int) *Clock {
	return &Clock{nodeID: nodeID, Counter: 0}
}

// LocalEvent records an internal action on this node only, incrementing its counter by 1.
func (clock *Clock) LocalEvent() TimeStamp {
	clock.lock.Lock()
	defer clock.lock.Unlock()

	clock.Counter++
	return TimeStamp{Counter: clock.Counter, NodeID: clock.nodeID}
}

// SendMessage increments counter for the send event and attaches the returned timestamp to the message.
func (clock *Clock) SendMessage() TimeStamp {
	return clock.LocalEvent()
}

// ReceiveMessage fast-forwards the local counter if incoming timestamp is ahead, then increments by 1.
func (clock *Clock) ReceiveMessage(senderKaTimeStamp TimeStamp) TimeStamp {
	clock.lock.Lock()
	defer clock.lock.Unlock()

	if senderKaTimeStamp.Counter > clock.Counter {
		clock.Counter = senderKaTimeStamp.Counter
	}

	clock.Counter++
	return TimeStamp{Counter: clock.Counter, NodeID: clock.nodeID}
}
