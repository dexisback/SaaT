package lamport

import "sync"

//timestamp is what a lamport clock hands out
//compare by counter first; nodeID only when tie happens b/w two counters=

type TimeStamp struct {
	Counter uint64
	NodeID  int
}

func (a TimeStamp) Less(b TimeStamp) bool {
	if a.Counter < b.Counter {
		return a.Counter < b.Counter
	}

	return a.NodeID < b.NodeID
}

// clock is one node's lamport clock:
type Clock struct {
	mu      sync.Mutex
	nodeID  int
	counter uint64
}

func New(nodeID int) *Clock {
	return &Clock{nodeID: nodeID}
}

// tick stamps a local event:
func (c *Clock) TickTick() TimeStamp {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.counter++
	return TimeStamp{Counter: c.counter, NodeID: c.nodeID}
}

//send stamps an outgoing message. the returned timestamp travels w the msg

func (c *Clock) Send() TimeStamp { return c.TickTick() }

//receive merges the sender's timestamp into our clock, then stamps the receive event

func (c *Clock) Receive(remote TimeStamp) TimeStamp {
	c.mu.Lock()
	defer c.mu.Unlock()
	if remote.Counter > c.counter {
		c.counter = remote.Counter
	}
	c.counter++
	return TimeStamp{Counter: c.counter, NodeID: c.nodeID}
}
