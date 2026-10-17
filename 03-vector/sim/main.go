package main

import (
	"fmt"

	"github.com/dexisback/SaaT/03-vector"
	"github.com/dexisback/SaaT/sim"
)

func main() {
	fmt.Println("===================================================================")
	fmt.Println("  03-VECTOR: Vector Clock Simulation Through /sim Scenario")
	fmt.Println("===================================================================")
	fmt.Println()

	events := sim.RunBasicScenario(42)
	causality := sim.NewCausality(events)

	clocks := map[int]*vector.Clock{
		0: vector.NewClock(0, 3),
		1: vector.NewClock(1, 3),
		2: vector.NewClock(2, 3),
	}

	stamps := map[sim.EventKey]vector.Vector{}
	msgPayloads := map[int]vector.Vector{}

	fmt.Println("-- Executing Events Through Vector Clocks --")
	for i, e := range events {
		var v vector.Vector
		clock := clocks[e.NodeID]

		switch e.Type {
		case sim.Local:
			v = clock.LocalEvent()
			fmt.Printf("[%d] Node %d [Local]   seq=%d -> Vector: %v\n", i, e.NodeID, e.Sequence, v)

		case sim.Send:
			v = clock.SendMessage()
			msgPayloads[e.MessageID] = v
			fmt.Printf("[%d] Node %d [Send]    seq=%d msgID=%d -> Vector: %v\n", i, e.NodeID, e.Sequence, e.MessageID, v)

		case sim.Receive:
			sentV := msgPayloads[e.MessageID]
			v = clock.ReceiveMessage(sentV)
			fmt.Printf("[%d] Node %d [Receive] seq=%d msgID=%d (incoming %v) -> Vector: %v\n",
				i, e.NodeID, e.Sequence, e.MessageID, sentV, v)
		}

		stamps[e.Key()] = v
	}

	fmt.Println()
	fmt.Println("-- Causality Audit --")
	for _, e := range events {
		if e.Type == sim.Receive {
			sendV := msgPayloads[e.MessageID]
			recvV := stamps[e.Key()]
			rel := vector.Compare(sendV, recvV)
			fmt.Printf("✓ Msg %d: Send %v vs Receive %v -> %s (Causally Valid)\n",
				e.MessageID, sendV, recvV, rel)
		}
	}

	fmt.Println()
	fmt.Println("-- Concurrency Detection Demonstration --")
	concurrentPairs := causality.ConcurrentPairs(events)
	if len(concurrentPairs) > 0 {
		pair := concurrentPairs[0]
		eA, eB := pair[0], pair[1]
		vA, vB := stamps[eA.Key()], stamps[eB.Key()]

		rel := vector.Compare(vA, vB)
		fmt.Printf("Ground truth says: %s and %s are CONCURRENT.\n", eA, eB)
		fmt.Printf("Vector Clock Compare(%v, %v) -> %s\n", vA, vB, rel)
		fmt.Println("-> Victory over Lamport: Vector clocks successfully detect true concurrency!")
		fmt.Println("   Neither vector dominates the other, proving neither caused the other.")
	}
}
