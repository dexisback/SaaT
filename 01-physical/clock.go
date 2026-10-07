// simulating a physical wall clock for a single node
package physical

import "time"

//clock simulates one node's wall clock
//offset is how far this node's clock is from "true" time
// a -ve offset means the clock runs behind, then we make it equal

type Clock struct {
	offset time.Duration
}

func New(offset time.Duration) *Clock {
	return &Clock{offset: offset}
}

//read returns what this node "thinks" the time is,
//given the actual true time

// so with Read(), we are simulating "drift".if we pass in the truetime into this function, it returns what the node thinks the time is by addingits custom offset
// For example, if trueTime is 12:00:00 and the node's clock has an offset of +5 seconds, Read() will return 12:00:05.
func (c *Clock) Read(trueTime time.Time) time.Time {
	return trueTime.Add(c.offset)
}

//step simulates an NTP adjustment:
//the clock is suddenly jerked by delta (negative = jumps backwards)
//cant really use real ntp here; refer ./realNTP/main.go

func (c *Clock) Step(delta time.Duration) {
	c.offset += delta
}

//NOTE: Read takes trueTime as a parameter instead of calling time.Now() inside. That is what makes the demo deterministic. We decide what "true time" is, and the clock only distorts it.
//Step is the NTP-style jump. Calling Step(-100 * time.Millisecond) between two events on the same node will produce the non-monotonic timestamps.
