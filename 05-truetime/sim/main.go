package main

import (
	"fmt"
	"sync/atomic"
	"time"

	"github.com/dexisback/SaaT/05-truetime"
	"github.com/dexisback/SaaT/sim"
)

func main() {
	fmt.Println("===================================================================")
	fmt.Println("  05-TRUETIME: TrueTime Uncertainty & Commit-Wait in /sim")
	fmt.Println("===================================================================")
	fmt.Println()

	events := sim.RunBasicScenario(42)

	var globalPhysicalTime int64 = 1000
	uncertainty := int64(7) // +/- 7ms uncertainty window

	// Background ticker simulating physical time passage
	stopTicker := make(chan struct{})
	go func() {
		ticker := time.NewTicker(1 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				atomic.AddInt64(&globalPhysicalTime, 1)
			case <-stopTicker:
				return
			}
		}
	}()
	defer close(stopTicker)

	clockNode0 := truetime.NewClock(uncertainty, func() int64 {
		return atomic.LoadInt64(&globalPhysicalTime)
	})
	clockNode1 := truetime.NewClock(uncertainty, func() int64 {
		return atomic.LoadInt64(&globalPhysicalTime) - 3 // Node 1 runs 3ms behind
	})

	fmt.Println("-- Simulated TrueTime States at t=1000ms --")
	fmt.Println("Node 0 Interval:", clockNode0.Now())
	fmt.Println("Node 1 Interval:", clockNode1.Now())

	fmt.Println()
	fmt.Println("-- Executing Scenario Messages with External Consistency --")
	var t1CommitStamp int64

	for i, e := range events {
		time.Sleep(5 * time.Millisecond)

		if e.NodeID == 0 && e.Type == sim.Send {
			// Transaction T1 commits on Node 0 before message is sent
			t1CommitStamp = clockNode0.Stamp()
			fmt.Printf("[%d] Node 0 commits Tx A with Stamp: %d (Latest bound)\n", i, t1CommitStamp)

			// 1. Without commit wait:
			isPast := clockNode1.After(t1CommitStamp)
			fmt.Printf("   WITHOUT wait -> Node 1 After(Tx A)? %v (Risk of external ordering inversion!)\n", isPast)

			// 2. With commit wait:
			fmt.Println("   WITH Commit-Wait -> Node 0 pauses until absolute time passes Stamp...")
			clockNode0.CommitWait(t1CommitStamp)
			time.Sleep(5 * time.Millisecond) // buffer for skewed node 1 to catch up

			isPastNow := clockNode1.After(t1CommitStamp)
			fmt.Printf("   AFTER wait  -> Node 1 After(Tx A)? %v (Guaranteed external consistency!)\n", isPastNow)
		}
	}

	fmt.Println()
	fmt.Println("-- Summary: Why Spanner Needed TrueTime --")
	fmt.Println("By enforcing commit-wait (waiting out 2*epsilon uncertainty), Google Spanner guarantees")
	fmt.Println("that if transaction T2 starts after T1 commits anywhere on Earth, T2's timestamp > T1's.")
}
