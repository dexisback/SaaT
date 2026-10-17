package main

import (
	"fmt"

	"github.com/dexisback/SaaT/09-interval-tree-clocks"
)

func main() {
	fmt.Println("👇🏼👇🏼👇🏼👇🏼 start: one node owns everything 👇🏼👇🏼👇🏼👇🏼")
	nodeA := itc.NewStamp()
	fmt.Println("node A:", nodeA)

	fmt.Println()
	fmt.Println("👇🏼👇🏼👇🏼👇🏼 fork: node A spawns node B 👇🏼👇🏼👇🏼👇🏼")
	nodeA, nodeB := nodeA.Fork()
	fmt.Println("node A after fork:", nodeA)
	fmt.Println("node B after fork:", nodeB)

	fmt.Println()
	fmt.Println("👇🏼👇🏼👇🏼👇🏼independent events 👇🏼👇🏼👇🏼👇🏼")
	nodeA.Event()
	fmt.Println("node A after local event:", nodeA)

	nodeB.Event()
	nodeB.Event()
	fmt.Println("node B after two local events:", nodeB)

	fmt.Println()
	fmt.Println("👇🏼👇🏼👇🏼👇🏼 fork again: node B spawns node C 👇🏼👇🏼👇🏼👇🏼")
	nodeB, nodeC := nodeB.Fork()
	fmt.Println("node B after second fork:", nodeB)
	fmt.Println("node C (new node):", nodeC)

	nodeC.Event()
	fmt.Println("node C after local event:", nodeC)

	fmt.Println()
	fmt.Println("👇🏼👇🏼👇🏼👇🏼join: node C shuts down, merges back into B 👇🏼👇🏼👇🏼👇🏼")
	nodeB = itc.Join(nodeB, nodeC)
	fmt.Println("node B after joining C:", nodeB)

	fmt.Println()
	fmt.Println("👇🏼👇🏼👇🏼👇🏼ordering check 👇🏼👇🏼👇🏼👇🏼")
	fmt.Println("node A happened before node B?", nodeA.HappenedBefore(nodeB))
	fmt.Println("node B happened before node A?", nodeB.HappenedBefore(nodeA))
}
