package hlc

import (
	"fmt"
	"sync"
)

// TimeStamp represents a Hybrid Logical Clock timestamp:
// 1. WallTime: Physical time in milliseconds, bounded to real physical time.
// 2. Counter: Logical counter to break ties when two events share the same WallTime.
type TimeStamp struct {
	WallTime int64
	Counter  int
}

// String formats a timestamp as (WallTime, Counter)
func (timestamp TimeStamp) String() string {
	return fmt.Sprintf("(%d, %d)", timestamp.WallTime, timestamp.Counter)
}

// IsBefore compares WallTime first, then uses Counter to break ties.
func (timestamp TimeStamp) IsBefore(other TimeStamp) bool {
	if timestamp.WallTime != other.WallTime {
		return timestamp.WallTime < other.WallTime
	}
	return timestamp.Counter < other.Counter
}

func max64(first int64, second int64) int64 {
	if first > second {
		return first
	}
	return second
}

// Clock is one node's Hybrid Logical Clock.
type Clock struct {
	lock     sync.Mutex
	wallTime int64
	counter  int
}

func NewClock() *Clock {
	return &Clock{wallTime: 0, counter: 0}
}

// LocalEvent records an internal event given the current physical time reading.
func (clock *Clock) LocalEvent(physicalTimeNow int64) TimeStamp {
	clock.lock.Lock()
	defer clock.lock.Unlock()

	previousWallTime := clock.wallTime
	clock.wallTime = max64(previousWallTime, physicalTimeNow)

	if clock.wallTime == previousWallTime {
		clock.counter++ // physical time hasn't moved past us, increment counter
	} else {
		clock.counter = 0 // physical time moved ahead, reset counter to 0
	}

	return TimeStamp{WallTime: clock.wallTime, Counter: clock.counter}
}

// SendMessage records a send event and returns the HLC timestamp attached to the message.
func (clock *Clock) SendMessage(physicalTimeNow int64) TimeStamp {
	return clock.LocalEvent(physicalTimeNow)
}

// ReceiveMessage updates the clock with an incoming message timestamp and local physical time reading.
func (clock *Clock) ReceiveMessage(incomingTimeStamp TimeStamp, physicalTimeNow int64) TimeStamp {
	clock.lock.Lock()
	defer clock.lock.Unlock()

	previousWallTime := clock.wallTime
	clock.wallTime = max64(max64(previousWallTime, incomingTimeStamp.WallTime), physicalTimeNow)

	sameAsMine := clock.wallTime == previousWallTime
	sameAsMessage := clock.wallTime == incomingTimeStamp.WallTime

	if sameAsMessage && sameAsMine {
		if incomingTimeStamp.Counter > clock.counter {
			clock.counter = incomingTimeStamp.Counter
		}
		clock.counter++
	} else if sameAsMine {
		clock.counter++
	} else if sameAsMessage {
		clock.counter = incomingTimeStamp.Counter + 1
	} else {
		clock.counter = 0
	}

	return TimeStamp{WallTime: clock.wallTime, Counter: clock.counter}
}
