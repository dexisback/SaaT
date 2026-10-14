package main

import (
	"fmt"
	"sync"
)

// a lamport timestamp is two numbers: the counter, and the id of the node that stamped it:
type TimeStamp struct {
	Counter int
	NodeID  int
}

//isBefore says whether one TimeStamp is ordered before another TimeStamp (same for isAfter )

//main primary comparison of who arrived first or later -> done by counters. if counters clash -> compare NodeIDs

func (first TimeStamp) isBefore(second TimeStamp) bool {
	if first.Counter != second.Counter {
		return first.Counter < second.Counter
	}
	// else:
	return first.NodeID < second.NodeID
}

//clock is one node's lamport clock. (each node has its own, and they never share memory)

type Clock struct {
	lock    sync.Mutex //protects counter if several goroutines are using this clock
	nodeID  int
	Counter int
}

func NewClock(nodeID int) *Clock {
	return &Clock{nodeID: nodeID, Counter: 0}
}

// localEvent: something happened on _this_ node ONLY. hence we add 1 to the counte r
func (clock *Clock) LocalEvent() TimeStamp {
	clock.lock.Lock()
	defer clock.lock.Unlock()

	clock.Counter++
	return TimeStamp{Counter: clock.Counter, NodeID: clock.nodeID}

}

// sendMessage: sending is also an event. The TimeStamp it returns is attached to the message, so the receiver can see it
func (clock *Clock) SendMessage() TimeStamp {
	return clock.LocalEvent()
}

// receiveMessage: a message arrives carrying the sender's timestamp -> jump our counter up to the sender's if it is ahead, then add 1
func (clock *Clock) ReceiveMessage(senderKaTimeStamp TimeStamp) TimeStamp {
	clock.lock.Lock()
	defer clock.lock.Unlock()

	if senderKaTimeStamp.Counter > clock.Counter {
		clock.Counter = senderKaTimeStamp.Counter //jump to match (DDIA)
	}

	//else it is alr equal:
	clock.Counter++
	return TimeStamp{Counter: clock.Counter, NodeID: clock.nodeID}

}

// main tells a small story w three nodes, one line per event:
func main() {
	node0Clock := NewClock(0)
	node1Clock := NewClock(1)
	node2Clock := NewClock(2)

	//node 0 does something locally, then sends a message to node 1
	node0Local := node0Clock.LocalEvent()
	node0Send := node0Clock.SendMessage()

	//node 1 receives that msg, then sends to node 2:
	node1Receive := node1Clock.ReceiveMessage(node0Send)
	node1Send := node1Clock.SendMessage()

	//node 2 does something local on its own, then receives node 1's message:
	node2Local := node2Clock.LocalEvent()
	node2Receive := node2Clock.ReceiveMessage(node1Send)

	//print dumpster:
	fmt.Println("node 0 local   :", node0Local.Counter)
	fmt.Println("node 0 send    :", node0Send.Counter)
	fmt.Println("node 1 receive :", node1Receive.Counter)
	fmt.Println("node 1 send    :", node1Send.Counter)
	fmt.Println("node 2 local   :", node2Local.Counter)
	fmt.Println("node 2 receive :", node2Receive.Counter)

	fmt.Println()
	fmt.Println("node 0 send is before node 1 receive?", node0Send.isBefore(node1Receive))
	fmt.Println("node 2 local is before node 0 send?  ", node2Local.isBefore(node0Send))
	fmt.Println("node 0 local is before node 2 local? ", node0Local.isBefore(node2Local))

}
