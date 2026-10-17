package main

import (
	"fmt"

	hv "github.com/dexisback/SaaT/11-hybrid-vector"
)

func main() {
	fmt.Println("=== same story as always, three nodes ===")
	fmt.Println()

	clockOfNode0 := hv.NewClock(0, 3)
	clockOfNode1 := hv.NewClock(1, 3)
	clockOfNode2 := hv.NewClock(2, 3)

	// Node 0: local event at wall time 1000, then sends at 1010
	node0Local := clockOfNode0.LocalEvent(1000)
	node0Send := clockOfNode0.SendMessage(1010)

	// Node 1: its clock is 50ms behind (wall time 980), receives then sends
	node1Receive := clockOfNode1.ReceiveMessage(node0Send, 980)
	node1Send := clockOfNode1.SendMessage(990)

	// Node 2: local event at 1005, then receives
	node2Local := clockOfNode2.LocalEvent(1005)
	node2Receive := clockOfNode2.ReceiveMessage(node1Send, 1065)

	fmt.Println("node 0 local   :", node0Local)
	fmt.Println("node 0 send    :", node0Send)
	fmt.Println("node 1 receive :", node1Receive, " <- node 1's clock said 980")
	fmt.Println("node 1 send    :", node1Send)
	fmt.Println("node 2 local   :", node2Local)
	fmt.Println("node 2 receive :", node2Receive)

	fmt.Println()
	fmt.Println("👇🏼👇🏼👇🏼👇🏼 comparison 👇🏼👇🏼👇🏼👇🏼")
	fmt.Println()
	fmt.Println("node 0 send vs node 1 receive:", hv.Compare(node0Send, node1Receive))
	fmt.Println("node 2 local vs node 0 send  :", hv.Compare(node2Local, node0Send))
	fmt.Println("node 0 local vs node 2 local :", hv.Compare(node0Local, node2Local))

	fmt.Println()
	fmt.Println("👇🏼👇🏼👇🏼👇🏼 what hybrid vector adds over plain vector 👇🏼👇🏼👇🏼👇🏼")
	fmt.Println()
	fmt.Println("each slot carries a wall time, so you can tell roughly WHEN")
	fmt.Println("an event happened on each node, not just how many events there were.")
	fmt.Println("node 1's slot in node 2's receive:", node2Receive[1])
	fmt.Println("that (1010, 2) means: node 1's last known event was around wall time 1010.")
	fmt.Println("a plain vector would only tell you: node 1 had 2 events.")
}
