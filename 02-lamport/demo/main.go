package main

import (
	"fmt"

	"github.com/dexisback/SaaT/02-lamport"
)

func main() {
	node0Clock := lamport.NewClock(0)
	node1Clock := lamport.NewClock(1)
	node2Clock := lamport.NewClock(2)

	// Node 0 does something locally, then sends a message to node 1
	node0Local := node0Clock.LocalEvent()
	node0Send := node0Clock.SendMessage()

	// Node 1 receives that msg, then sends to node 2
	node1Receive := node1Clock.ReceiveMessage(node0Send)
	node1Send := node1Clock.SendMessage()

	// Node 2 does something local on its own, then receives node 1's message
	node2Local := node2Clock.LocalEvent()
	node2Receive := node2Clock.ReceiveMessage(node1Send)

	fmt.Println("node 0 local   :", node0Local.Counter)
	fmt.Println("node 0 send    :", node0Send.Counter)
	fmt.Println("node 1 receive :", node1Receive.Counter)
	fmt.Println("node 1 send    :", node1Send.Counter)
	fmt.Println("node 2 local   :", node2Local.Counter)
	fmt.Println("node 2 receive :", node2Receive.Counter)

	fmt.Println()
	fmt.Println("node 0 send is before node 1 receive?", node0Send.IsBefore(node1Receive))
	fmt.Println("node 2 local is before node 0 send?  ", node2Local.IsBefore(node0Send))
	fmt.Println("node 0 local is before node 2 local? ", node0Local.IsBefore(node2Local))
}
