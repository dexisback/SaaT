package main

import (
	"fmt"

	"github.com/dexisback/SaaT/05-truetime"
)

func main() {
	currentTime := int64(1000)

	clockOfNode0 := truetime.NewClock(2, func() int64 { return currentTime })
	clockOfNode1 := truetime.NewClock(5, func() int64 { return currentTime - 3 })

	fmt.Println("=== what each node thinks the time is ===")
	fmt.Println("node 0 Now():", clockOfNode0.Now())
	fmt.Println("node 1 Now():", clockOfNode1.Now())

	fmt.Println()
	fmt.Println("=== transaction A commits on node 0 ===")
	timestampA := clockOfNode0.Stamp()
	fmt.Println("node 0 stamps transaction A at:", timestampA, "(Latest end of uncertainty)")

	fmt.Println()
	fmt.Println("=== what happens WITHOUT commit wait ===")
	fmt.Println("without waiting, node 1's Earliest is", clockOfNode1.Now().Earliest,
		"and A's timestamp is", timestampA)
	fmt.Println("is node 1 certain that transaction A happened before now?",
		clockOfNode1.After(timestampA))
	fmt.Println("-> Node 1 cannot be certain A is in the past! It sits inside node 1's uncertainty window.")
	fmt.Println("-> Without commit wait, two transactions on different nodes can appear to swap order.")

	fmt.Println()
	fmt.Println("=== what happens WITH commit wait ===")
	currentTime = 1008 // time moves forward
	fmt.Println("time advances to 1008 ms...")
	fmt.Println("node 0 commits after waiting until its Earliest >", timestampA)
	clockOfNode0.CommitWait(timestampA)
	fmt.Println("node 0 has finished commit wait. Transaction A is now externally visible.")

	fmt.Println()
	fmt.Println("now node 1 checks after the wait:")
	fmt.Println("node 1's Earliest is now", clockOfNode1.Now().Earliest)
	fmt.Println("is node 1 certain that transaction A happened before now?",
		clockOfNode1.After(timestampA))
	fmt.Println("-> Node 1 is now CERTAIN! Transaction A is strictly in node 1's past.")
}
