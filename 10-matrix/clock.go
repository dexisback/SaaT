//vector clock : each node keeps one counter per node. A vector of size N
//matrix clock: each node keeps ONE VECTOR PER NODE. (An NxN grid). Row "i" is what node "i" thinks the vector clock looks like rn
//why? cuz it answers "has every node in the system seen at least this event?"
//why is this needed? for garbage collecting old log entries . once every node ocnfirms it has seen an event, you can safely delete it. matrix clocks give you that confirmation without an extra round of messages

package main

import "fmt"

// Matrix is an N×N grid.
// matrix[i][j] = "what node i thinks node j's counter is"
// Your own row (matrix[myID]) is your vector clock.
// Other rows tell you what those nodes last told you about their view.

type Matrix [][]int

func newMatrix(numberOfNodes int) Matrix {
	matrix := make(Matrix, numberOfNodes)
	for rowIndex := range matrix {
		matrix[rowIndex] = make([]int, numberOfNodes)
	}

	return matrix
}

func copyMatrix(original Matrix) Matrix {
	numberOfNodes := len(original)
	copied := newMatrix(numberOfNodes)
	for rowIndex := range original {
		copy(copied[rowIndex], original[rowIndex])
	}
	return copied
}

func rowWiseMax(firstRow []int, secondRow []int) []int {
	result := make([]int, len(firstRow))
	for slot := range firstRow {
		if firstRow[slot] > secondRow[slot] {
			result[slot] = firstRow[slot]
		} else {
			result[slot] = secondRow[slot]
		}
	}

	return result
}

// clock is one node's matrix clock:
type Clock struct {
	nodeID        int
	numberOfNodes int
	matrix        Matrix
}

func NewClock(nodeID int, numberOfNodes int) *Clock {
	return &Clock{
		nodeID:        nodeID,
		numberOfNodes: numberOfNodes,
		matrix:        newMatrix(numberOfNodes),
	}
}

//localevent: bump our own slot in our own row then update every other row's entry for us with our new counter

func (clock *Clock) LocalEvent() Matrix {
	clock.matrix[clock.nodeID][clock.nodeID]++

	// tell every row what our new counter is
	ourNewCounter := clock.matrix[clock.nodeID][clock.nodeID]
	for rowIndex := range clock.matrix {
		if clock.matrix[rowIndex][clock.nodeID] < ourNewCounter {
			clock.matrix[rowIndex][clock.nodeID] = ourNewCounter
		}
	}
	return copyMatrix(clock.matrix)
}

// sendMeddage: same as a local event.the entire matrix travels with the message
func (clock *Clock) SendMessage() Matrix {
	return clock.LocalEvent()
}

//receiveMessage: merge incoming matrix with ours row by row (slot-wise max), then do a local event to record the receive itself

func (clock *Clock) ReceiveMessage(incomingMatrix Matrix) Matrix {
	for rowIndex := range clock.matrix {
		clock.matrix[rowIndex] = rowWiseMax(clock.matrix[rowIndex], incomingMatrix[rowIndex])
	}
	return clock.LocalEvent()
}

// canDelete: if every node has seen atleast counter events from sourceNode, then anything upto and including that counter can be garbage collected
func (clock *Clock) CanDelete(sourceNode int, counter int) bool {
	for rowIndex := range clock.matrix {
		if clock.matrix[rowIndex][sourceNode] < counter {
			return false // this row (some node) hasn't confirmed seeing that event yet
		}
	}
	return true
}

func printMatrix(label string, matrix Matrix) {
	fmt.Println(label)
	for rowIndex, row := range matrix {
		fmt.Printf(" row %d: %v\n", rowIndex, row)
	}
}

func main() {
	fmt.Println("👇🏼👇🏼👇🏼👇🏼 setup: three nodes 👇🏼👇🏼👇🏼👇🏼")
	fmt.Println()

	clockOfNode0 := NewClock(0, 3)
	clockOfNode1 := NewClock(1, 3)
	clockOfNode2 := NewClock(2, 3)

	// node 0 does a local event then sends to node 1
	node0Local := clockOfNode0.LocalEvent()
	node0Send := clockOfNode0.SendMessage()

	fmt.Println("node 0 after local event and send:")
	printMatrix("  matrix:", node0Local)
	_ = node0Send

	// node 1 receives from node 0, then sends to node 2
	node1Receive := clockOfNode1.ReceiveMessage(node0Send)
	node1Send := clockOfNode1.SendMessage()

	fmt.Println("node 1 after receiving from node 0:")
	printMatrix("  matrix:", node1Receive)
	_ = node1Send

	// node 2 does its own local event, then receives from node 1
	clockOfNode2.LocalEvent()
	node2Receive := clockOfNode2.ReceiveMessage(node1Send)

	fmt.Println("node 2 after receiving from node 1:")
	printMatrix("  matrix:", node2Receive)

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
	fmt.Printf("row 0 (what node 2 thinks node 0 has seen): %v\n", clockOfNode2.matrix[0])
	fmt.Printf("row 1 (what node 2 thinks node 1 has seen): %v\n", clockOfNode2.matrix[1])
	fmt.Printf("row 2 (what node 2 thinks node 2 has seen): %v\n", clockOfNode2.matrix[2])
}
