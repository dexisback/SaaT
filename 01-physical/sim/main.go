package main

import (
	"fmt"
	"time"

	physical "github.com/dexisback/SaaT/01-physical"
	"github.com/dexisback/SaaT/sim"
)

func main() {
	fmt.Println("===================================================================")
	fmt.Println("  01-PHYSICAL: Wall Clock Simulation Through /sim Scenario")
	fmt.Println("===================================================================")
	fmt.Println()

	events := sim.RunBasicScenario(42)
	c := sim.NewCausality(events)

	// Simulated skew: Node 1 runs 50ms behind, Node 2 runs 5ms ahead
	offsets := map[int]time.Duration{
		0: 0,
		1: -50 * time.Millisecond,
		2: 5 * time.Millisecond,
	}

	// NTP backward step on Node 1
	steps := map[physical.EventKey]time.Duration{
		{Node: 1, Sequence: 1}: -100 * time.Millisecond,
	}

	stamped := physical.Stamp(events, offsets, steps)
	physical.Report(stamped)

	fmt.Println()
	fmt.Println("-- Causality DAG Comparison --")
	fmt.Println("Ground truth concurrent pairs:", len(c.ConcurrentPairs(events)))
	fmt.Println("Verdict: Physical clocks fail causality when skew exceeds network latency,")
	fmt.Println("and backward NTP steps break local monotonicity.")
}
