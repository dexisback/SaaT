package lamport

import "sync"

//timestamp is what a lamport clock hands out
//compare by counter first; nodeID only when tie happens b/w two counters=

type Timestamp struct {
	Counter uint64
	NodeID  int
}

// TimeStamp is retained as an alias for Timestamp for backwards compatibility.
type TimeStamp = Timestamp

func (a Timestamp) Less(b Timestamp) bool {
	if a.Counter != b.Counter {
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
func (c *Clock) TickTick() Timestamp {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.counter++
	return Timestamp{Counter: c.counter, NodeID: c.nodeID}
}

// Tick is an alias for TickTick.
func (c *Clock) Tick() Timestamp { return c.TickTick() }

//send stamps an outgoing message. the returned timestamp travels w the msg

func (c *Clock) Send() Timestamp { return c.TickTick() }

//receive merges the sender's timestamp into our clock, then stamps the receive event

func (c *Clock) Receive(remote Timestamp) Timestamp {
	c.mu.Lock()
	defer c.mu.Unlock()
	if remote.Counter > c.counter {
		c.counter = remote.Counter
	}
	c.counter++
	return Timestamp{Counter: c.counter, NodeID: c.nodeID}
}
