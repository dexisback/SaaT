package main

import (
	"fmt"

	"github.com/dexisback/SaaT/09-interval-tree-clocks"
	"github.com/dexisback/SaaT/sim"
)

func main() {
	fmt.Println("===================================================================")
	fmt.Println("  09-ITC: Dynamic Membership & Causal Ordering in /sim")
	fmt.Println("===================================================================")
	fmt.Println()

	// ITC solves dynamic membership: nodes can join and leave without central registration
	events := sim.RunBasicScenario(42)

	fmt.Println("-- Simulating Dynamic Node Topology with Interval Tree Clocks --")
	// Start with a single root node owning the entire [0, 1] identity space
	nodeA := itc.NewStamp()
	fmt.Println("Initial Root Node A owning entire space:", nodeA)

	// Node A dynamically forks Node B (splits identity in half)
	nodeA, nodeB := nodeA.Fork()
	fmt.Println("Node A forks Node B:")
	fmt.Println("  Node A ID:", nodeA)
	fmt.Println("  Node B ID:", nodeB)

	// Replay local and send events
	stamps := map[sim.EventKey]*itc.Stamp{}
	stamps[sim.EventKey{Node: 0, Sequence: 0}] = itc.NewStamp()

	fmt.Println()
	fmt.Println("-- Executing Scenario Events with ITC Stamps --")
	for i, e := range events {
		switch e.NodeID {
		case 0:
			nodeA.Event()
			fmt.Printf("[%d] Node A Event (seq=%d) -> %s\n", i, e.Sequence, nodeA)
		case 1:
			nodeB.Event()
			fmt.Printf("[%d] Node B Event (seq=%d) -> %s\n", i, e.Sequence, nodeB)
		case 2:
			// Node B dynamically forks Node C on the fly!
			var nodeC *itc.Stamp
			nodeB, nodeC = nodeB.Fork()
			nodeC.Event()
			fmt.Printf("[%d] Node C (Dynamic Child) Event (seq=%d) -> %s\n", i, e.Sequence, nodeC)

			// Node C joins back into Node B
			nodeB = itc.Join(nodeB, nodeC)
			fmt.Printf("    -> Node C merges back into Node B -> %s\n", nodeB)
		}
	}

	fmt.Println()
	fmt.Println("-- Causality & Identity Conservation Invariant --")
	// Merging Node A and Node B back restores the entire [0, 1] space
	recombined := itc.Join(nodeA, nodeB)
	fmt.Println("Recombined Root after all node joins:", recombined)
	fmt.Println()
	fmt.Println("Verdict: Unlike Vector Clocks (which require fixed N node slots),")
	fmt.Println("Interval Tree Clocks grow and shrink dynamically without coordinators or wasted slots.")
}
