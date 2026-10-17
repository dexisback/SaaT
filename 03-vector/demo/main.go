package main

import (
	"fmt"

	"github.com/dexisback/SaaT/03-vector"
)

func main() {
	clockOfNode0 := vector.NewClock(0, 3)
	clockOfNode1 := vector.NewClock(1, 3)
	clockOfNode2 := vector.NewClock(2, 3)

	// Node 0 does something locally, then sends a message to node 1
	node0Local := clockOfNode0.LocalEvent()
	node0Send := clockOfNode0.SendMessage()

	// Node 1 receives that message, then sends one to node 2
	node1Receive := clockOfNode1.ReceiveMessage(node0Send)
	node1Send := clockOfNode1.SendMessage()

	// Node 2 does something on its own, then receives node 1's message
	node2Local := clockOfNode2.LocalEvent()
	node2Receive := clockOfNode2.ReceiveMessage(node1Send)

	fmt.Println("node 0 local   :", node0Local)
	fmt.Println("node 0 send    :", node0Send)
	fmt.Println("node 1 receive :", node1Receive)
	fmt.Println("node 1 send    :", node1Send)
	fmt.Println("node 2 local   :", node2Local)
	fmt.Println("node 2 receive :", node2Receive)

	fmt.Println()
	fmt.Println("node 0 send  vs node 1 receive:", vector.Compare(node0Send, node1Receive))
	fmt.Println("node 2 local vs node 0 send   :", vector.Compare(node2Local, node0Send))
	fmt.Println("node 0 local vs node 2 local  :", vector.Compare(node0Local, node2Local))
	fmt.Println("node 0 send  vs node 2 receive:", vector.Compare(node0Send, node2Receive))
}
