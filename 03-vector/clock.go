package main

import (
	"fmt"
	"sync"
)

//a vector has one slot per node .
//slot 0 = "how many events of node 0 do i know about"
//slot 1 = "how many events of node 1 do i know about" , and so on....

type Vector []int

//copyOf makes a sep copy of a vector. why? cuz in go , a slice is a window onto shared memory . it we handed out the clock's own evector, later events would silently change timestamps we already stored

func copyof(OGVector Vector) Vector {
	copiedVector := make(Vector, len(OGVector))
	copy(copiedVector, OGVector)
	return copiedVector
}

//compare says how to vectors relate (before/after/equal/concurrent)

func Compare(first Vector, second Vector) string {
	firstHasSmallerSlot := false //init
	secondHasSmallerSlot := false

	for slot := range first {
		if first[slot] < second[slot] {
			firstHasSmallerSlot = true
		}
		//else if :
		if first[slot] > second[slot] {
			secondHasSmallerSlot = true

		}
	}

	if firstHasSmallerSlot && secondHasSmallerSlot {
		return "concurrent"
	}
	if firstHasSmallerSlot {
		return "before"
	}
	if secondHasSmallerSlot {
		return "after"
	}
	return "equal"
}

// clock is one node's vector clock , similar to 01 and 02
type Clock struct {
	lock   sync.Mutex
	nodeID int
	vector Vector
}

// NewClock
func NewClock(nodeID int, numberOfNodes int) *Clock {
	return &Clock{nodeID: nodeID, vector: make(Vector, numberOfNodes)}
}

// localEvent: something happened only on this node
func (clock *Clock) LocalEvent() Vector {
	clock.lock.Lock()
	defer clock.lock.Unlock()

	clock.vector[clock.nodeID]++ //increase our own counter
	return copyof(clock.vector)
}

// sendMessage: sending is also an event.the returned vector is attached to the msg so the receiver can see it
func (clock *Clock) SendMessage() Vector {
	return clock.LocalEvent()
}

//receiveMessage: a message arrived carrying the sender's vector
//for every slot, we keep the bigger number , then add 1 to our own slot.

func (clock *Clock) ReceiveMessage(incomingVector Vector) Vector {
	clock.lock.Lock()
	defer clock.lock.Unlock()
	for slot := range clock.vector {
		if incomingVector[slot] > clock.vector[slot] {
			clock.vector[slot] = incomingVector[slot]
		}
	}

	clock.vector[clock.nodeID]++
	return copyof(clock.vector)

}

func main() {
	clockOfNode0 := NewClock(0, 3)
	clockOfNode1 := NewClock(1, 3)
	clockOfNode2 := NewClock(2, 3)

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
	fmt.Println("node 0 send  vs node 1 receive:", Compare(node0Send, node1Receive))
	fmt.Println("node 2 local vs node 0 send   :", Compare(node2Local, node0Send))
	fmt.Println("node 0 local vs node 2 local  :", Compare(node0Local, node2Local))
	fmt.Println("node 0 send  vs node 2 receive:", Compare(node0Send, node2Receive))
}

//lamport couldnt recognise the node 2 and node 0 as concurrent. lamport ranked those same pairs as "before", and
//everyone starts with [0 0 0], and there are only two rules;
//1: local event or send : add +1 to your own slot
//2: receive: for each slot, keep the bigger of (your number, the number in the msg). Then do +1 to your own slot, (because the receive itself is a new event)

//[events on node 0, events on node1, events on node2]
//for example: node 0 local -> [1 0 0 ]. then node 0 sends to node 1 => [2 0 0 ]. then node 1 receives the msg => [2 1 0] (bigger of 0 and 2 is 2), and so on
