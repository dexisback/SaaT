package main

import (
	"fmt"

	"github.com/dexisback/SaaT/07-dotted-version-vector"
	"github.com/dexisback/SaaT/sim"
)

func main() {
	fmt.Println("===================================================================")
	fmt.Println("  07-DOTTED-VERSION-VECTOR: Write Identity & Sync in /sim")
	fmt.Println("===================================================================")
	fmt.Println()

	events := sim.RunPartitionAndHeal(42)

	rep0 := dvv.NewReplica(0, 2)
	rep1 := dvv.NewReplica(1, 2)

	// Step 1: Initial shared state
	rep0.Write("init_doc", dvv.NewVersionVector(2))
	rep1.Sync(rep0)
	fmt.Printf("Initial shared state: Value=%q Dot=%v Vector=%v\n\n",
		rep0.Current.Value, rep0.Current.Dot, rep0.VersionVector)

	// Step 2: Clients read during partition
	_, clientCtx0 := rep0.Read()
	_, clientCtx1 := rep1.Read()

	fmt.Println("-- Executing Partitioned Writes with Dots --")
	for i, e := range events {
		switch e.Type {
		case sim.Local:
			if e.NodeID == 0 {
				rep0.Write("doc_edited_by_clientA", clientCtx0)
				fmt.Printf("[%d] Replica 0 writes: Value=%q Dot=%v Context=%v\n",
					i, rep0.Current.Value, rep0.Current.Dot, rep0.Current.KnownVersions)
			} else {
				rep1.Write("doc_edited_by_clientB", clientCtx1)
				fmt.Printf("[%d] Replica 1 writes: Value=%q Dot=%v Context=%v\n",
					i, rep1.Current.Value, rep1.Current.Dot, rep1.Current.KnownVersions)
			}

		case sim.Receive:
			if e.NodeID == 1 {
				hasSeen := rep0.Current.KnownVersions.HasSeen(rep1.Current.Dot)
				fmt.Printf("\n[%d] Checking causal relationship before sync:\n", i)
				fmt.Printf("    Did Replica 0's client see Replica 1's dot %v? %v\n", rep1.Current.Dot, hasSeen)
				res := rep1.Sync(rep0)
				fmt.Printf("    Replica 1 Sync result: %s\n", res)
			}
		}
	}

	fmt.Println()
	fmt.Println("-- Summary: Dotted Version Vectors --")
	fmt.Println("Dots decouple the write event identity from the causality context, preventing silent")
	fmt.Println("overwrites and allowing distributed databases like Riak to preserve concurrent versions.")
}
