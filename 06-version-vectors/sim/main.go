package main

import (
	"fmt"

	vv "github.com/dexisback/SaaT/06-version-vectors"
	"github.com/dexisback/SaaT/sim"
)

func main() {
	fmt.Println("===================================================================")
	fmt.Println("  06-VERSION-VECTORS: Replica Partition & Sync in /sim")
	fmt.Println("===================================================================")
	fmt.Println()

	// Replicas experience a partition, do independent writes, then heal & sync
	events := sim.RunPartitionAndHeal(42)

	rep0 := vv.NewReplica(0, 2)
	rep1 := vv.NewReplica(1, 2)

	fmt.Println("-- Executing Partition & Sync Scenario --")
	for i, e := range events {
		switch e.Type {
		case sim.Local:
			if e.NodeID == 0 {
				val := fmt.Sprintf("val_node0_rev%d", e.Sequence)
				rep0.Write(val)
				fmt.Printf("[%d] Replica 0 [Local Write] -> Value=%q Vector=%v\n", i, rep0.Value, rep0.VersionVector)
			} else {
				val := fmt.Sprintf("val_node1_rev%d", e.Sequence)
				rep1.Write(val)
				fmt.Printf("[%d] Replica 1 [Local Write] -> Value=%q Vector=%v\n", i, rep1.Value, rep1.VersionVector)
			}

		case sim.Send:
			fmt.Printf("[%d] Replica %d [Send Sync Msg %d] -> Preparing to reconcile\n", i, e.NodeID, e.MessageID)

		case sim.Receive:
			if e.NodeID == 1 {
				// Replica 1 attempts to sync Replica 0's diverged write
				res := rep1.Sync(rep0.Value, rep0.VersionVector)
				fmt.Printf("[%d] Replica 1 [Sync From Rep 0] -> %s\n", i, res)
				fmt.Printf("    -> Divergence detected! Resolving with application merge...\n")
				rep1.ResolveConflict("val_merged(node0+node1)", rep0.VersionVector)
				fmt.Printf("    -> Resolved: Value=%q Vector=%v\n", rep1.Value, rep1.VersionVector)
			} else if e.NodeID == 0 {
				// Replica 0 receives Replica 1's resolved write
				res := rep0.Sync(rep1.Value, rep1.VersionVector)
				fmt.Printf("[%d] Replica 0 [Sync From Rep 1] -> %s\n", i, res)
				fmt.Printf("    -> Replicas converge! Value=%q Vector=%v\n", rep0.Value, rep0.VersionVector)
			}
		}
	}

	fmt.Println()
	fmt.Println("-- Summary: Version Vectors in Dynamo-Style Storage --")
	fmt.Println("Version vectors allow replicas to independently accept writes during network partitions,")
	fmt.Println("detect causal divergence, and merge cleanly once the partition heals.")
}
