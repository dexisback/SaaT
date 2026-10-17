package truetime

import (
	"fmt"
	"sync"
	"time"
)

// Interval represents TrueTime's [Earliest, Latest] time interval.
type Interval struct {
	Earliest int64
	Latest   int64
}

func (interval Interval) String() string {
	return fmt.Sprintf("[%d, %d] (uncertainty: %d ms)", interval.Earliest, interval.Latest, interval.Latest-interval.Earliest)
}

// Clock simulates a TrueTime API providing uncertainty intervals and commit wait.
type Clock struct {
	lock            sync.Mutex
	uncertainty     int64
	physicalTimeNow func() int64
}

func NewClock(uncertainty int64, physicalTimeNow func() int64) *Clock {
	if physicalTimeNow == nil {
		physicalTimeNow = func() int64 { return time.Now().UnixMilli() }
	}

	return &Clock{
		uncertainty:     uncertainty,
		physicalTimeNow: physicalTimeNow,
	}
}

// Stamp picks the Latest bound of Now().
// Conservative stamping ensures that once any node's Earliest passes it,
// the transaction is guaranteed to be in the absolute real-time past.
func (clock *Clock) Stamp() int64 {
	return clock.Now().Latest
}

// Now returns the uncertainty interval [midpoint - uncertainty, midpoint + uncertainty].
func (clock *Clock) Now() Interval {
	clock.lock.Lock()
	defer clock.lock.Unlock()

	midpoint := clock.physicalTimeNow()
	return Interval{
		Earliest: midpoint - clock.uncertainty,
		Latest:   midpoint + clock.uncertainty,
	}
}

// CommitWait blocks until the clock's Earliest bound is strictly greater than the stamped timestamp.
func (clock *Clock) CommitWait(timestamp int64) {
	for {
		now := clock.Now()
		if now.Earliest > timestamp {
			return
		}
		time.Sleep(1 * time.Millisecond)
	}
}

// After returns true if the clock is certain the current time is after timestamp.
func (clock *Clock) After(timestamp int64) bool {
	return clock.Now().Earliest > timestamp
}