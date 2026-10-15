//gives a range. an interval. a confidence  . wohi spanner wali bkc
//google's spanner does this w gps receivers and atomic clocks in every datacenter. the hardware tells each machine ki bhai yeh hai "earliest" and yeh hai "latest"
//spanner then uses this to make a guarantee no other db makes:"if transaction A commits before transaction B, A's timestamp is strictly smaller than B's"

//commit wait -- before a tx can commit, it must wait till earlier > latest_of_what_it_stamped. which makes sense cuh we wanna be absolutely clear and sure ki that event happened so we wait for the window to be over

//-------------_NOTE: we cannot replicate GPS hardware, but we can simulate here the uncertainity level. each node's "clock" returns an internal, and commits wait it out

package main

import (
	"fmt"
	"sync"
	"time"
)

type Interval struct {
	Earliest int64
	Latest   int64
}

func (interval Interval) String() string {
	return fmt.Sprintf("[%d, %d] (uncertainty: %d ms)", interval.Earliest, interval.Latest, interval.Latest-interval.Earliest)
}

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
// We pick the latest possible time on purpose: we want to be conservative,
// claiming the transaction happened as late as possible, so any reader can
// be sure it's in the past once their Earliest passes it.
func (clock *Clock) Stamp() int64 {
	return clock.Now().Latest
}

func (clock *Clock) Now() Interval {
	clock.lock.Lock()
	defer clock.lock.Unlock()

	midpoint := clock.physicalTimeNow()
	return Interval{
		Earliest: midpoint - clock.uncertainty,
		Latest:   midpoint + clock.uncertainty,
	}
}

func (clock *Clock) CommitWait(timestamp int64) {
	for {
		now := clock.Now()
		if now.Earliest > timestamp {
			return
		}
		time.Sleep(1 * time.Millisecond)
	}
}

func (clock *Clock) After(timestamp int64) bool {
	return clock.Now().Earliest > timestamp
}

func main() {
	currentTime := int64(1000)

	clockOfNode0 := NewClock(2, func() int64 { return currentTime })
	clockOfNode1 := NewClock(5, func() int64 { return currentTime - 3 })

	fmt.Println("=== what each node thinks the time is ===")
	fmt.Println("node 0 Now():", clockOfNode0.Now())
	fmt.Println("node 1 Now():", clockOfNode1.Now())

	fmt.Println()
	fmt.Println("=== transaction A commits on node 0 ===")
	timestampA := clockOfNode0.Stamp()
	fmt.Println("node 0 stamps transaction A at:", timestampA, "(Latest end of uncertainty)")

	fmt.Println()
	fmt.Println("=== what happens WITHOUT commit wait ===")
	fmt.Println("without waiting, node 1's Earliest is", clockOfNode1.Now().Earliest,
		"and A's timestamp is", timestampA)
	fmt.Println("is node 1 certain that transaction A happened before now?",
		clockOfNode1.After(timestampA))
	fmt.Println("-> Node 1 cannot be certain A is in the past! It sits inside node 1's uncertainty window.")
	fmt.Println("-> Without commit wait, two transactions on different nodes can appear to swap order.")

	fmt.Println()
	fmt.Println("=== WITH commit wait on node 0 ===")
	// In production with GPS/atomic clocks, this is where the node sleeps.
	// In this simulated story, we advance the simulated clock to represent the wait:
	currentTime = 1015
	fmt.Println("node 0 waits out the uncertainty window...")
	clockOfNode0.CommitWait(timestampA)
	fmt.Println("node 0 Earliest is now", clockOfNode0.Now().Earliest,
		"- certain timestamp", timestampA, "is now in the past worldwide")

	fmt.Println()
	fmt.Println("=== transaction B starts on node 1 (after A's commit wait) ===")
	fmt.Println("node 1 Now():", clockOfNode1.Now())
	fmt.Println("is node 1 certain that transaction A happened before now?",
		clockOfNode1.After(timestampA))
	timestampB := clockOfNode1.Stamp()
	fmt.Println("node 1 stamps transaction B at:", timestampB)
	fmt.Println("external consistency guaranteed: A's stamp (", timestampA, ") < B's stamp (", timestampB, ")")
}









/*
Summary / How to read it:

1. Node 0's interval is [998, 1002]. Its physical clock says 1000, and uncertainty is ±2.
   It cannot claim the real time is exactly 1000.

2. Transaction A is stamped at 1002, the Latest end.
   We pick the latest possible time on purpose: we want to be conservative, claiming the transaction
   happened as late as possible, so any reader can be sure it's in the past once their Earliest passes it.

3. Without Commit Wait:
   Node 1's Earliest is 992, which is less than A's timestamp of 1002.
   1002 sits inside Node 1's uncertainty window: maybe A already happened, maybe it hasn't.
   Without commit wait, transactions on different nodes can appear to swap order!

4. With Commit Wait:
   Node 0 must wait until its own Earliest > 1002 (here fast-forwarded to Earliest=1013).
   Because real time >= Earliest, real time across the entire universe is now guaranteed to be > 1002.
   Only then does Node 0 release locks and reply to the client.

5. When Transaction B starts on Node 1:
   Node 1's Earliest is now 1007 (> 1002).
   Node 1 is 100% certain A happened in the past, and stamps B at 1017 (> 1002).
   External consistency (linearizability) is guaranteed without cross-datacenter messaging!
*/