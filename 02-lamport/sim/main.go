package main

import (
	"fmt"

	"github.com/dexisback/SaaT/02-lamport"
	"github.com/dexisback/SaaT/sim"
)

func main() {
	fmt.Println("===================================================================")
	fmt.Println("  02-LAMPORT: Logical Clock Simulation Through /sim Scenario")
	fmt.Println("===================================================================")
	fmt.Println()

	events := sim.RunBasicScenario(42)
	causality := sim.NewCausality(events)

	clocks := map[int]*lamport.Clock{
		0: lamport.NewClock(0),
		1: lamport.NewClock(1),
		2: lamport.NewClock(2),
	}

	stamps := map[sim.EventKey]lamport.TimeStamp{}
	msgPayloads := map[int]lamport.TimeStamp{}

	fmt.Println("-- Executing Events Through Lamport Clocks --")
	for i, e := range events {
		var ts lamport.TimeStamp
		clock := clocks[e.NodeID]

		switch e.Type {
		case sim.Local:
			ts = clock.LocalEvent()
			fmt.Printf("[%d] Node %d [Local]   seq=%d -> Lamport: (counter=%d, node=%d)\n",
				i, e.NodeID, e.Sequence, ts.Counter, ts.NodeID)

		case sim.Send:
			ts = clock.SendMessage()
			msgPayloads[e.MessageID] = ts
			fmt.Printf("[%d] Node %d [Send]    seq=%d msgID=%d -> Lamport: (counter=%d, node=%d)\n",
				i, e.NodeID, e.Sequence, e.MessageID, ts.Counter, ts.NodeID)

		case sim.Receive:
			sentTS := msgPayloads[e.MessageID]
			ts = clock.ReceiveMessage(sentTS)
			fmt.Printf("[%d] Node %d [Receive] seq=%d msgID=%d (incoming counter=%d) -> Lamport: (counter=%d, node=%d)\n",
				i, e.NodeID, e.Sequence, e.MessageID, sentTS.Counter, ts.Counter, ts.NodeID)
		}

		stamps[e.Key()] = ts
	}

	fmt.Println()
	fmt.Println("-- Causality & Monotonicity Audit --")
	violations := 0
	for _, e := range events {
		if e.Sequence > 0 {
			prevKey := sim.EventKey{Node: e.NodeID, Sequence: e.Sequence - 1}
			if !stamps[prevKey].IsBefore(stamps[e.Key()]) {
				fmt.Printf("VIOLATION: local sequence broken on node %d\n", e.NodeID)
				violations++
			}
		}
	}

	// Verify message causal edges
	for _, e := range events {
		if e.Type == sim.Receive {
			sendTS := msgPayloads[e.MessageID]
			recvTS := stamps[e.Key()]
			if !sendTS.IsBefore(recvTS) {
				fmt.Printf("VIOLATION: message %d send TS %v not before receive TS %v\n", e.MessageID, sendTS, recvTS)
				violations++
			} else {
				fmt.Printf("✓ Msg %d: Send (counter=%d) -> Receive (counter=%d) correctly ordered\n",
					e.MessageID, sendTS.Counter, recvTS.Counter)
			}
		}
	}

	if violations == 0 {
		fmt.Println("Result: 0 Causal Inversions. Lamport preserves happened-before perfectly!")
	}

	fmt.Println()
	fmt.Println("-- Concurrency Blindspot Demonstration --")
	concurrentPairs := causality.ConcurrentPairs(events)
	if len(concurrentPairs) > 0 {
		pair := concurrentPairs[0]
		eA, eB := pair[0], pair[1]
		tsA, tsB := stamps[eA.Key()], stamps[eB.Key()]

		fmt.Printf("Ground truth says: %s and %s are CONCURRENT.\n", eA, eB)
		fmt.Printf("Lamport says: %s isBefore %s? %v\n", eA, eB, tsA.IsBefore(tsB))
		fmt.Println("-> Insight: Lamport clocks CANNOT detect concurrency. They impose an arbitrary")
		fmt.Println("   total order using NodeID tie-breaking even when events are causally independent.")
	}
}
