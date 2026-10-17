package main

import (
	"fmt"

	"github.com/dexisback/SaaT/10-matrix"
)

func main() {
	fmt.Println("👇🏼👇🏼👇🏼👇🏼 setup: three nodes 👇🏼👇🏼👇🏼👇🏼")
	fmt.Println()

	clockOfNode0 := matrix.NewClock(0, 3)
	clockOfNode1 := matrix.NewClock(1, 3)
	clockOfNode2 := matrix.NewClock(2, 3)

	// Node 0 does a local event then sends to node 1
	node0Local := clockOfNode0.LocalEvent()
	node0Send := clockOfNode0.SendMessage()

	fmt.Println("node 0 after local event and send:")
	matrix.PrintMatrix("  matrix:", node0Local)
	_ = node0Send

	// Node 1 receives from node 0, then sends to node 2
	node1Receive := clockOfNode1.ReceiveMessage(node0Send)
	node1Send := clockOfNode1.SendMessage()

	fmt.Println("node 1 after receiving from node 0:")
	matrix.PrintMatrix("  matrix:", node1Receive)
	_ = node1Send

	// Node 2 does its own local event, then receives from node 1
	clockOfNode2.LocalEvent()
	node2Receive := clockOfNode2.ReceiveMessage(node1Send)

	fmt.Println("node 2 after receiving from node 1:")
	matrix.PrintMatrix("  matrix:", node2Receive)

	fmt.Println()
	fmt.Println("👇🏼👇🏼👇🏼👇🏼garbage collection check 👇🏼👇🏼👇🏼👇🏼")
	fmt.Println()
	fmt.Println("can node 2 delete node 0's first event (counter=1)?",
		clockOfNode2.CanDelete(0, 1))
	fmt.Println("can node 2 delete node 0's second event (counter=2)?",
		clockOfNode2.CanDelete(0, 2))
	fmt.Println()
	fmt.Println("why: node 2's matrix shows what every node has seen.")
	fmt.Println("if every row's slot for node 0 is >= the counter, it's safe to delete.")

	fmt.Println()
	fmt.Println("👇🏼👇🏼👇🏼👇🏼 what each row means in node 2's final matrix 👇🏼👇🏼👇🏼👇🏼")
	fmt.Println()
	m2 := clockOfNode2.Matrix()
	fmt.Printf("row 0 (what node 2 thinks node 0 has seen): %v\n", m2[0])
	fmt.Printf("row 1 (what node 2 thinks node 1 has seen): %v\n", m2[1])
	fmt.Printf("row 2 (what node 2 thinks node 2 has seen): %v\n", m2[2])
}
