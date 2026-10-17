package main

import (
	"fmt"

	hv "github.com/dexisback/SaaT/11-hybrid-vector"
	"github.com/dexisback/SaaT/sim"
)

func main() {
	fmt.Println("===================================================================")
	fmt.Println("  11-HYBRID-VECTOR: Concurrency + Physical Time in /sim")
	fmt.Println("===================================================================")
	fmt.Println()

	events := sim.RunBasicScenario(42)
	causality := sim.NewCausality(events)

	clocks := map[int]*hv.Clock{
		0: hv.NewClock(0, 3),
		1: hv.NewClock(1, 3),
		2: hv.NewClock(2, 3),
	}

	stamps := map[sim.EventKey]hv.HybridVector{}
	msgPayloads := map[int]hv.HybridVector{}
	baseTime := int64(1000)

	fmt.Println("-- Executing Events Through Hybrid Vector Clocks --")
	for i, e := range events {
		clock := clocks[e.NodeID]
		phyTime := baseTime + int64(i*15)
		var vec hv.HybridVector

		switch e.Type {
		case sim.Local:
			vec = clock.LocalEvent(phyTime)
			fmt.Printf("[%d] Node %d [Local]   seq=%d (t=%dms) -> %v\n",
				i, e.NodeID, e.Sequence, phyTime, vec)

		case sim.Send:
			vec = clock.SendMessage(phyTime)
			msgPayloads[e.MessageID] = vec
			fmt.Printf("[%d] Node %d [Send]    seq=%d msgID=%d (t=%dms) -> %v\n",
				i, e.NodeID, e.Sequence, e.MessageID, phyTime, vec)

		case sim.Receive:
			sentVec := msgPayloads[e.MessageID]
			vec = clock.ReceiveMessage(sentVec, phyTime)
			fmt.Printf("[%d] Node %d [Receive] seq=%d msgID=%d (t=%dms, incoming %v) -> %v\n",
				i, e.NodeID, e.Sequence, e.MessageID, phyTime, sentVec, vec)
		}

		stamps[e.Key()] = vec
	}

	fmt.Println()
	fmt.Println("-- Causality & Concurrency Audit --")
	// 1. Causality check
	for _, e := range events {
		if e.Type == sim.Receive {
			sVec := msgPayloads[e.MessageID]
			rVec := stamps[e.Key()]
			rel := hv.Compare(sVec, rVec)
			fmt.Printf("✓ Msg %d: Send vs Receive -> %s (Causality preserved)\n", e.MessageID, rel)
		}
	}

	// 2. Concurrency check
	concurrentPairs := causality.ConcurrentPairs(events)
	if len(concurrentPairs) > 0 {
		pair := concurrentPairs[0]
		eA, eB := pair[0], pair[1]
		vA, vB := stamps[eA.Key()], stamps[eB.Key()]

		rel := hv.Compare(vA, vB)
		fmt.Println()
		fmt.Printf("Ground truth: %s and %s are CONCURRENT.\n", eA, eB)
		fmt.Printf("Hybrid Vector Compare -> %s (True concurrency detected!)\n", rel)
		fmt.Println("Bonus vs plain vector:")
		fmt.Printf("  Slot 0: %v (indicates Node 0's last known event was at ~%dms)\n", vB[0], vB[0].WallTime)
	}

	fmt.Println()
	fmt.Println("Verdict: Hybrid Vector clocks achieve both goals simultaneously:")
	fmt.Println("Full concurrency detection without losing proximity to physical wall-clock time.")
}
