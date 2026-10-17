package matrix

import "fmt"

// Matrix is an N×N grid.
// matrix[i][j] = "what node i thinks node j's counter is"
// Your own row (matrix[myID]) is your vector clock.
// Other rows tell you what those nodes last told you about their view.
type Matrix [][]int

func NewMatrix(numberOfNodes int) Matrix {
	matrix := make(Matrix, numberOfNodes)
	for rowIndex := range matrix {
		matrix[rowIndex] = make([]int, numberOfNodes)
	}
	return matrix
}

func CopyMatrix(original Matrix) Matrix {
	numberOfNodes := len(original)
	copied := NewMatrix(numberOfNodes)
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

// Clock is one node's matrix clock.
type Clock struct {
	NodeID        int
	NumberOfNodes int
	matrix        Matrix
}

func NewClock(nodeID int, numberOfNodes int) *Clock {
	return &Clock{
		NodeID:        nodeID,
		NumberOfNodes: numberOfNodes,
		matrix:        NewMatrix(numberOfNodes),
	}
}

// Matrix returns a defensive copy of the current matrix state.
func (clock *Clock) Matrix() Matrix {
	return CopyMatrix(clock.matrix)
}

// LocalEvent bumps our own slot in our own row.
func (clock *Clock) LocalEvent() Matrix {
	clock.matrix[clock.NodeID][clock.NodeID]++
	return CopyMatrix(clock.matrix)
}

// SendMessage records a send event and returns the matrix to attach to the message.
func (clock *Clock) SendMessage() Matrix {
	return clock.LocalEvent()
}

// ReceiveMessage updates our knowledge from the incoming matrix.
func (clock *Clock) ReceiveMessage(incomingMatrix Matrix) Matrix {
	// 1. Update our knowledge of what other nodes have seen
	for rowIndex := range clock.matrix {
		clock.matrix[rowIndex] = rowWiseMax(clock.matrix[rowIndex], incomingMatrix[rowIndex])
	}

	// 2. Our own row (our vector clock) must reflect all events observed by the sender
	for _, row := range incomingMatrix {
		for colIndex := range clock.matrix[clock.NodeID] {
			if row[colIndex] > clock.matrix[clock.NodeID][colIndex] {
				clock.matrix[clock.NodeID][colIndex] = row[colIndex]
			}
		}
	}

	return clock.LocalEvent()
}

// CanDelete returns true if every node has confirmed seeing at least counter events from sourceNode.
func (clock *Clock) CanDelete(sourceNode int, counter int) bool {
	for rowIndex := range clock.matrix {
		if clock.matrix[rowIndex][sourceNode] < counter {
			return false
		}
	}
	return true
}

func PrintMatrix(label string, matrix Matrix) {
	fmt.Println(label)
	for rowIndex, row := range matrix {
		fmt.Printf(" row %d: %v\n", rowIndex, row)
	}
}
