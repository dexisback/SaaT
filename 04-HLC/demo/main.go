package main

import (
	"fmt"

	hlc "github.com/dexisback/SaaT/04-HLC"
)

func main() {
	clockOfNode0 := hlc.NewClock()
	clockOfNode1 := hlc.NewClock()
	clockOfNode2 := hlc.NewClock()

	// Node 0 does something locally, then sends a message to node 1
	node0Local := clockOfNode0.LocalEvent(1000)
	node0Send := clockOfNode0.SendMessage(1010)

	// Node 1 receives that message (its slow clock says 980), then sends one to node 2
	node1Receive := clockOfNode1.ReceiveMessage(node0Send, 980)
	node1Send := clockOfNode1.SendMessage(990)

	// Node 2 does something on its own, then receives node 1's message
	node2Local := clockOfNode2.LocalEvent(1005)
	node2Receive := clockOfNode2.ReceiveMessage(node1Send, 1065)

	fmt.Println("node 0 local   :", node0Local)
	fmt.Println("node 0 send    :", node0Send)
	fmt.Println("node 1 receive :", node1Receive, " <- its own clock said 980")
	fmt.Println("node 1 send    :", node1Send)
	fmt.Println("node 2 local   :", node2Local)
	fmt.Println("node 2 receive :", node2Receive)

	fmt.Println()
	fmt.Println("node 0 send is before node 1 receive?", node0Send.IsBefore(node1Receive))
	fmt.Println("node 1 send is before node 2 receive?", node1Send.IsBefore(node2Receive))
	fmt.Println("node 2 local is before node 0 send?  ", node2Local.IsBefore(node0Send))
}
