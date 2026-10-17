package main

import (
	"fmt"
	"time"

	physical "github.com/dexisback/SaaT/01-physical"
	"github.com/dexisback/SaaT/sim"
)

func main() {
	events := sim.RunBasicScenario(69)

	offsets := map[int]time.Duration{
		0: 0,
		1: -50 * time.Millisecond, // node 1 runs 50ms behind
		2: 5 * time.Millisecond,   // node 2 runs 5ms ahead
	}

	fmt.Println("=== Scenario A: Clock Skew Only ===")
	physical.Report(physical.Stamp(events, offsets, nil))

	fmt.Println()
	fmt.Println("=== Scenario B: Skew + Negative NTP Step Jump on Node 1 ===")
	steps := map[physical.EventKey]time.Duration{
		{Node: 1, Sequence: 1}: -100 * time.Millisecond, // clock jumps back before node 1's second event
	}
	physical.Report(physical.Stamp(events, offsets, steps))
}
