package main

import (
	"fmt"

	"github.com/dexisback/SaaT/07-dotted-version-vector"
)

func main() {
	fmt.Println("=== the scenario that breaks plain version vectors ===")
	fmt.Println()

	replicaR := dvv.NewReplica(0, 2)

	// Initial write to get things started
	replicaR.Write("v1", dvv.NewVersionVector(2))
	fmt.Println("replica writes v1, vector:", replicaR.VersionVector)

	// Client 1 and Client 2 both read the same version
	_, contextSeenByClient1 := replicaR.Read()
	_, contextSeenByClient2 := replicaR.Read()
	fmt.Println("client 1 reads, sees context:", contextSeenByClient1)
	fmt.Println("client 2 reads, sees context:", contextSeenByClient2)

	// Client 1 writes v2 on top of what it saw
	replicaR.Write("v2", contextSeenByClient1)
	dotV2 := replicaR.Current.Dot
	fmt.Println()
	fmt.Println("client 1 writes v2, vector:", replicaR.VersionVector)
	fmt.Println("current dot:", replicaR.Current.Dot)

	// Client 2 writes v3 on top of what it saw (the old context)
	replicaR.Write("v3", contextSeenByClient2)
	fmt.Println()
	fmt.Println("client 2 writes v3 (with stale context), vector:", replicaR.VersionVector)
	fmt.Println("current dot:", replicaR.Current.Dot)
	fmt.Println("current value:", replicaR.Current.Value)
	fmt.Println("client 2's known versions when it wrote:", replicaR.Current.KnownVersions)
	fmt.Println("did client 2 see v2 before writing v3?", replicaR.Current.KnownVersions.HasSeen(dotV2))

	fmt.Println()
	fmt.Println("=== normal sync: two replicas, one is ahead ===")
	fmt.Println()

	replicaA := dvv.NewReplica(0, 2)
	replicaB := dvv.NewReplica(1, 2)

	replicaA.Write("hello", dvv.NewVersionVector(2))
	fmt.Println("replica A writes 'hello', vector:", replicaA.VersionVector, "dot:", replicaA.Current.Dot)

	result := replicaB.Sync(replicaA)
	fmt.Println("replica B syncs from A:", result)
	fmt.Println("replica B value:", replicaB.Current.Value, "vector:", replicaB.VersionVector)

	replicaA.Write("hello-updated", dvv.CopyOf(replicaA.VersionVector))
	fmt.Println()
	fmt.Println("replica A updates to 'hello-updated', dot:", replicaA.Current.Dot)

	result = replicaB.Sync(replicaA)
	fmt.Println("replica B syncs from A again:", result)
	fmt.Println("replica B value:", replicaB.Current.Value, "vector:", replicaB.VersionVector)

	// Sync again to show idempotency
	result = replicaB.Sync(replicaA)
	fmt.Println("replica B syncs from A third time (idempotent):", result)
}
