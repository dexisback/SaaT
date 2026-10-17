package main

import (
	"fmt"

	hlc "github.com/dexisback/SaaT/04-HLC"
	"github.com/dexisback/SaaT/sim"
)

func main() {
	fmt.Println("===================================================================")
	fmt.Println("  04-HLC: Hybrid Logical Clock Simulation Through /sim Scenario")
	fmt.Println("===================================================================")
	fmt.Println()

	events := sim.RunBasicScenario(42)
	causality := sim.NewCausality(events)

	clocks := map[int]*hlc.Clock{
		0: hlc.NewClock(),
		1: hlc.NewClock(),
		2: hlc.NewClock(),
	}

	// Simulated physical clocks per node: Node 1 runs 50ms behind Node 0
	nodePhysicalOffsets := map[int]int64{
		0: 0,
		1: -50, // Node 1 lagging by 50ms!
		2: 10,  // Node 2 slightly ahead
	}

	basePhysicalTime := int64(1000)
	stamps := map[sim.EventKey]hlc.TimeStamp{}
	msgPayloads := map[int]hlc.TimeStamp{}

	fmt.Println("-- Executing Events Through HLC (with simulated clock skew) --")
	for i, e := range events {
		clock := clocks[e.NodeID]
		phyTime := basePhysicalTime + int64(i*10) + nodePhysicalOffsets[e.NodeID]
		var ts hlc.TimeStamp

		switch e.Type {
		case sim.Local:
			ts = clock.LocalEvent(phyTime)
			fmt.Printf("[%d] Node %d [Local]   seq=%d (phy=%dms) -> HLC: %v\n",
				i, e.NodeID, e.Sequence, phyTime, ts)

		case sim.Send:
			ts = clock.SendMessage(phyTime)
			msgPayloads[e.MessageID] = ts
			fmt.Printf("[%d] Node %d [Send]    seq=%d msgID=%d (phy=%dms) -> HLC: %v\n",
				i, e.NodeID, e.Sequence, e.MessageID, phyTime, ts)

		case sim.Receive:
			sentTS := msgPayloads[e.MessageID]
			ts = clock.ReceiveMessage(sentTS, phyTime)
			fmt.Printf("[%d] Node %d [Receive] seq=%d msgID=%d (phy=%dms, incoming %v) -> HLC: %v\n",
				i, e.NodeID, e.Sequence, e.MessageID, phyTime, sentTS, ts)
		}

		stamps[e.Key()] = ts
	}

	fmt.Println()
	fmt.Println("-- Skew Resilience & Causality Audit --")
	for _, e := range events {
		if e.Type == sim.Receive {
			sendTS := msgPayloads[e.MessageID]
			recvTS := stamps[e.Key()]
			isOrdered := sendTS.IsBefore(recvTS)
			fmt.Printf("✓ Msg %d: Send HLC %v < Receive HLC %v -> %v (Even though receiver was 50ms slow!)\n",
				e.MessageID, sendTS, recvTS, isOrdered)
		}
	}

	fmt.Println()
	fmt.Println("-- The Tradeoff vs Vector Clocks --")
	concurrentPairs := causality.ConcurrentPairs(events)
	if len(concurrentPairs) > 0 {
		pair := concurrentPairs[0]
		eA, eB := pair[0], pair[1]
		tsA, tsB := stamps[eA.Key()], stamps[eB.Key()]

		fmt.Printf("Ground truth: %s and %s are CONCURRENT.\n", eA, eB)
		fmt.Printf("HLC verdict : %s IsBefore %s? %v\n", eA, eB, tsA.IsBefore(tsB))
		fmt.Println("-> Insight: HLC fixes physical clock skew without growing in size like vector clocks,")
		fmt.Println("   but like Lamport, it orders concurrent events rather than detecting them.")
	}
}
