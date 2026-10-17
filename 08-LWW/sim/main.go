package main

import (
	"fmt"
	"time"

	"github.com/dexisback/SaaT/08-LWW"
	"github.com/dexisback/SaaT/sim"
)

func main() {
	fmt.Println("===================================================================")
	fmt.Println("  08-LWW: Last-Write-Wins Register Simulation Through /sim")
	fmt.Println("===================================================================")
	fmt.Println()

	// Scenario where Node 0 and Node 2 write concurrently and race to Node 1
	events := sim.RunConcurrentRace(42)

	wallReg := lww.NewWallRegister()
	lamportReg := lww.NewLamportRegister()

	baseTime := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)

	// Simulated wall clock skews: Node 2's clock is 100ms behind Node 0
	nodeSkews := map[int]time.Duration{
		0: 0,
		1: 0,
		2: -100 * time.Millisecond, // Skewed backwards!
	}

	fmt.Println("-- Executing Concurrent Writes Across Replicas --")
	for i, e := range events {
		if e.Type == sim.Local {
			val := fmt.Sprintf("val_node%d_seq%d", e.NodeID, e.Sequence)

			// 1. Wall Clock Write
			wallTS := baseTime.Add(time.Duration(i*10)*time.Millisecond + nodeSkews[e.NodeID])
			wallOK := wallReg.Write(val, wallTS)

			// 2. Lamport Write
			lampTS := lww.LamportTimestamp{Counter: e.Sequence + 1, NodeID: e.NodeID}
			lampOK := lamportReg.Write(val, lampTS)

			fmt.Printf("[%d] Node %d Writes %-18q\n", i, e.NodeID, val)
			fmt.Printf("    -> Wall-Clock (t=%s): Accepted=%v Current=%q\n",
				wallTS.Format("15:04:05.000"), wallOK, wallReg.Read().Value)
			fmt.Printf("    -> Lamport    (%v)  : Accepted=%v Current=%q\n",
				lampTS, lampOK, lamportReg.Read().Value)
		}
	}

	fmt.Println()
	fmt.Println("-- Comparison: Wall-Clock LWW vs Lamport LWW --")
	fmt.Printf("Final Wall-Clock Value: %q (Notice: Node 2's write was dropped due to clock skew!)\n", wallReg.Read().Value)
	fmt.Printf("Final Lamport    Value: %q (Deterministic tie-breaking preserves latest logical write)\n", lamportReg.Read().Value)
	fmt.Println()
	fmt.Println("Verdict: Cassandra-style Wall-Clock LWW is prone to silent data loss under clock drift.")
	fmt.Println("Lamport LWW guarantees convergence across all nodes, but may arbitrarily discard writes.")
}
