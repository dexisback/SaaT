package main

import (
	"fmt"

	vv "github.com/dexisback/SaaT/06-version-vectors"
)

func main() {
	fmt.Println("👇🏼👇🏼👇🏼👇🏼 scenario 1: simple sequential write and sync 👇🏼👇🏼👇🏼👇🏼")
	repA := vv.NewReplica(0, 3)
	repB := vv.NewReplica(1, 3)

	repA.Write("version 1")
	fmt.Println("replica A writes:", repA.Value, "vector:", repA.VersionVector)

	result := repB.Sync(repA.Value, repA.VersionVector)
	fmt.Println("replica B syncs from A:", result)
	fmt.Println("replica B now has:", repB.Value, "vector:", repB.VersionVector)

	fmt.Println()
	fmt.Println("👇🏼👇🏼👇🏼👇🏼 scenario 2: concurrent writes (divergence) 👇🏼👇🏼👇🏼👇🏼")
	repA.Write("apple")
	fmt.Println("replica A writes 'apple':", repA.VersionVector)

	repB.Write("banana")
	fmt.Println("replica B writes 'banana':", repB.VersionVector)

	fmt.Println()
	fmt.Println("how does A relate to B?", vv.Compare(repA.VersionVector, repB.VersionVector))
	fmt.Println("how does B relate to A?", vv.Compare(repB.VersionVector, repA.VersionVector))

	fmt.Println()
	conflictResult := repB.Sync(repA.Value, repA.VersionVector)
	fmt.Println("replica B tries to sync from A:")
	fmt.Println(conflictResult)

	fmt.Println()
	fmt.Println("👇🏼👇🏼👇🏼👇🏼 resolving the conflict 👇🏼👇🏼👇🏼👇🏼")
	repB.ResolveConflict("fruit-salad", repA.VersionVector)
	fmt.Println("replica B resolves to 'fruit-salad', vector:", repB.VersionVector)

	syncAfterResolve := repA.Sync(repB.Value, repB.VersionVector)
	fmt.Println("replica A syncs from B after resolution:", syncAfterResolve)
	fmt.Println("replica A now has:", repA.Value, "vector:", repA.VersionVector)
}
