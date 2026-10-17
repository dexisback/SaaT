package main

import (
	"fmt"

	"github.com/dexisback/SaaT/10-matrix"
	"github.com/dexisback/SaaT/sim"
)

func main() {
	fmt.Println("===================================================================")
	fmt.Println("  10-MATRIX: Matrix Clock & Distributed GC in /sim")
	fmt.Println("===================================================================")
	fmt.Println()

	// 4-node relay: 0 -> 1 -> 2 -> 3
	events := sim.RunCausalChain(42)

	clocks := map[int]*matrix.Clock{
		0: matrix.NewClock(0, 4),
		1: matrix.NewClock(1, 4),
		2: matrix.NewClock(2, 4),
		3: matrix.NewClock(3, 4),
	}

	msgPayloads := map[int]matrix.Matrix{}

	fmt.Println("-- Executing 4-Node Causal Relay --")
	for i, e := range events {
		clock := clocks[e.NodeID]

		switch e.Type {
		case sim.Local:
			clock.LocalEvent()
			fmt.Printf("[%d] Node %d [Local]   seq=%d -> Counter: %d\n",
				i, e.NodeID, e.Sequence, clock.Matrix()[e.NodeID][e.NodeID])

		case sim.Send:
			m := clock.SendMessage()
			msgPayloads[e.MessageID] = m
			fmt.Printf("[%d] Node %d [Send]    seq=%d msgID=%d -> Transmitting Matrix\n",
				i, e.NodeID, e.Sequence, e.MessageID)

		case sim.Receive:
			incoming := msgPayloads[e.MessageID]
			clock.ReceiveMessage(incoming)
			fmt.Printf("[%d] Node %d [Receive] seq=%d msgID=%d -> Merged Knowledge\n",
				i, e.NodeID, e.Sequence, e.MessageID)
		}
	}

	fmt.Println()
	fmt.Println("-- Distributed Garbage Collection Audit (CanDelete) --")
	fmt.Println("Question: Can Node 0 safely delete its initial log entry (counter=1)?")

	for nodeID := 0; nodeID < 4; nodeID++ {
		canDel := clocks[nodeID].CanDelete(0, 1)
		fmt.Printf("Node %d view: CanDelete(Node 0, Event 1)? %v\n", nodeID, canDel)
	}

	fmt.Println()
	fmt.Println("-- Node 3 Final Matrix State --")
	matrix.PrintMatrix("Node 3 Knowledge Grid:", clocks[3].Matrix())

	fmt.Println()
	fmt.Println("Verdict: Matrix clocks tell each node not just what it has seen,")
	fmt.Println("but what EVERY OTHER NODE in the cluster has confirmed seeing,")
	fmt.Println("enabling safe distributed log compaction without extra coordination rounds.")
}
