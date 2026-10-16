package main

import "fmt"

//difference? replace each slot in a vector clock with an HLC timestamp . (wallTime, counter) instead of a plain integer.
//with this you get both concurrency detection + prximity to realtime

//merge rule is same as HLC but applied per slot (take the bigger walltime and if tie happens then the bigger counter wins)

//instead of a plain integer, each slot is an HLC timestamp

type HLCSlot struct {
	WallTime int64
	Counter  int
}

func (slot HLCSlot) String() string {
	return fmt.Sprintf("(%d,%d)", slot.WallTime, slot.Counter)
}


// isNewerThan compares two slots: wall time first, counter breaks ties.
func (slot HLCSlot) isNewerThan(other HLCSlot) bool {
	if slot.WallTime != other.WallTime {
		return slot.WallTime > other.WallTime
	}
	//else toh woh equal hi hai
	return slot.Counter > other.Counter
}


//mergeSlots takes the maximum of two HLC slots:
func mergeSlots(firstSlot HLCSlot, secondSlot HLCSlot) HLCSlot {
	if firstSlot.isNewerThan(secondSlot) {
		return firstSlot
	}
	return secondSlot
}



type HybridVector []HLCSlot     //one node's hybrid vector clock


func newHybridVector(numberOfNodes int) HybridVector {
	return make(HybridVector, numberOfNodes)
}

func copyHybridVector(original HybridVector) HybridVector {
	copied := make(HybridVector, len(original))
	copy(copied, original)
	return copied
}


//clock is one node's hybrid vector clock
type Clock struct {
	nodeID        int
	numberOfNodes int
	vector        HybridVector
}

func NewClock(nodeID int, numberOfNodes int) *Clock {
	return &Clock{
		nodeID:        nodeID,
		numberOfNodes: numberOfNodes,
		vector:        newHybridVector(numberOfNodes),
	}
}

//tickOwnSlot increases this node's own slot using the physical time . 
//same rule as HLC: take the bigger of (our current wall time, physical now), then if wall time didnt move, increment counter, elese reset counter to 0

func (clock *Clock) tickOwnSlot(physicalTimeNow int64) {
	currentSlot := clock.vector[clock.nodeID]
	newWallTime := currentSlot.WallTime
	if physicalTimeNow > newWallTime {
		newWallTime = physicalTimeNow
	}

	if newWallTime == currentSlot.WallTime {
		clock.vector[clock.nodeID] = HLCSlot{WallTime: newWallTime, Counter: currentSlot.Counter + 1}
	} else {
		clock.vector[clock.nodeID] = HLCSlot{WallTime: newWallTime, Counter: 0}
	}
}


// LocalEvent: tick our own slot, leave others unchanged.

func (clock *Clock) LocalEvent(physicalTimeNow int64) HybridVector {
	clock.tickOwnSlot(physicalTimeNow)
	return copyHybridVector(clock.vector)
}



//SendMsg: same as other sendMessages 
func (clock *Clock) SendMessage(physicalTimeNow int64) HybridVector {
	return clock.LocalEvent(physicalTimeNow)
}


//ReceiveMsg: merge incoming vector slot by slot (taking the max of each), then tick our own slot
func (clock *Clock) ReceiveMessage(incomingVector HybridVector, physicalTimeNow int64) HybridVector {
	for slot := range clock.vector {
		clock.vector[slot] = mergeSlots(clock.vector[slot], incomingVector[slot])
	}
	clock.tickOwnSlot(physicalTimeNow)
	return copyHybridVector(clock.vector)
}


type Relation string


const (
	Before     Relation = "before"
	After      Relation = "after"
	Equal      Relation = "equal"
	Concurrent Relation = "concurrent"
)

//Compare checks how firstVector relates to secondVector (but the thing is ki -- a slot in firstVector is "smaller" if its wall time is less. if walltime is equal then counter compare jiska less ho)

func Compare(firstVector HybridVector, secondVector HybridVector) Relation {
	firstHasSmallerSlot := false
	firstHasLargerSlot := false

	for slot := range firstVector {
		firstSlot := firstVector[slot]
		secondSlot := secondVector[slot]

		if secondSlot.isNewerThan(firstSlot) {
			firstHasSmallerSlot = true
		}
		if firstSlot.isNewerThan(secondSlot) {
			firstHasLargerSlot = true
		}
	}

	if firstHasSmallerSlot && firstHasLargerSlot {
		return Concurrent
	}
	if firstHasSmallerSlot {
		return Before
	}
	if firstHasLargerSlot {
		return After
	}
	return Equal
}





func main() {
	fmt.Println("=== same story as always, three nodes ===")
	fmt.Println()

	clockOfNode0 := NewClock(0, 3)
	clockOfNode1 := NewClock(1, 3)
	clockOfNode2 := NewClock(2, 3)

	// node 0: local event at wall time 1000, then sends at 1010
	node0Local := clockOfNode0.LocalEvent(1000)
	node0Send := clockOfNode0.SendMessage(1010)

	// node 1: its clock is 50ms behind (wall time 980), receives then sends
	node1Receive := clockOfNode1.ReceiveMessage(node0Send, 980)
	node1Send := clockOfNode1.SendMessage(990)

	// node 2: local event at 1005, then receives
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
	fmt.Println("node 0 send vs node 1 receive:", Compare(node0Send, node1Receive))
	fmt.Println("node 2 local vs node 0 send  :", Compare(node2Local, node0Send))
	fmt.Println("node 0 local vs node 2 local :", Compare(node0Local, node2Local))

	fmt.Println()
	fmt.Println("👇🏼👇🏼👇🏼👇🏼 what hybrid vector adds over plain vector 👇🏼👇🏼👇🏼👇🏼")
	fmt.Println()
	fmt.Println("each slot carries a wall time, so you can tell roughly WHEN")
	fmt.Println("an event happened on each node, not just how many events there were.")
	fmt.Println("node 1's slot in node 2's receive:", node2Receive[1])
	fmt.Println("that (1010, 2) means: node 1's last known event was around wall time 1010.")
	fmt.Println("a plain vector would only tell you: node 1 had 2 events.")
}




