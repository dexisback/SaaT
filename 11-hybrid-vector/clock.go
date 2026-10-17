package hybridvector

import "fmt"

// HLCSlot stores an HLC timestamp (WallTime, Counter) in each slot of a vector clock.
type HLCSlot struct {
	WallTime int64
	Counter  int
}

func (slot HLCSlot) String() string {
	return fmt.Sprintf("(%d,%d)", slot.WallTime, slot.Counter)
}

// IsNewerThan compares two slots: wall time first, counter breaks ties.
func (slot HLCSlot) IsNewerThan(other HLCSlot) bool {
	if slot.WallTime != other.WallTime {
		return slot.WallTime > other.WallTime
	}
	return slot.Counter > other.Counter
}

func (slot HLCSlot) isNewerThan(other HLCSlot) bool {
	return slot.IsNewerThan(other)
}

// MergeSlots takes the maximum of two HLC slots.
func MergeSlots(firstSlot HLCSlot, secondSlot HLCSlot) HLCSlot {
	if firstSlot.IsNewerThan(secondSlot) {
		return firstSlot
	}
	return secondSlot
}

func mergeSlots(firstSlot HLCSlot, secondSlot HLCSlot) HLCSlot {
	return MergeSlots(firstSlot, secondSlot)
}

// HybridVector represents one node's hybrid vector clock.
type HybridVector []HLCSlot

func NewHybridVector(numberOfNodes int) HybridVector {
	return make(HybridVector, numberOfNodes)
}

func CopyHybridVector(original HybridVector) HybridVector {
	copied := make(HybridVector, len(original))
	copy(copied, original)
	return copied
}

func copyHybridVector(original HybridVector) HybridVector {
	return CopyHybridVector(original)
}

// Clock is one node's hybrid vector clock.
type Clock struct {
	nodeID        int
	numberOfNodes int
	vector        HybridVector
}

func NewClock(nodeID int, numberOfNodes int) *Clock {
	return &Clock{
		nodeID:        nodeID,
		numberOfNodes: numberOfNodes,
		vector:        NewHybridVector(numberOfNodes),
	}
}

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

// LocalEvent ticks our own slot with physical time, leaving others unchanged.
func (clock *Clock) LocalEvent(physicalTimeNow int64) HybridVector {
	clock.tickOwnSlot(physicalTimeNow)
	return CopyHybridVector(clock.vector)
}

// SendMessage records a send event and returns the hybrid vector timestamp.
func (clock *Clock) SendMessage(physicalTimeNow int64) HybridVector {
	return clock.LocalEvent(physicalTimeNow)
}

// ReceiveMessage merges the incoming vector slot by slot, then ticks our own slot.
func (clock *Clock) ReceiveMessage(incomingVector HybridVector, physicalTimeNow int64) HybridVector {
	for slot := range clock.vector {
		clock.vector[slot] = MergeSlots(clock.vector[slot], incomingVector[slot])
	}
	clock.tickOwnSlot(physicalTimeNow)
	return CopyHybridVector(clock.vector)
}

type Relation string

const (
	Before     Relation = "before"
	After      Relation = "after"
	Equal      Relation = "equal"
	Concurrent Relation = "concurrent"
)

// Compare determines the causal ordering between two hybrid vectors.
func Compare(firstVector HybridVector, secondVector HybridVector) Relation {
	firstHasSmallerSlot := false
	firstHasLargerSlot := false

	for slot := range firstVector {
		firstSlot := firstVector[slot]
		secondSlot := secondVector[slot]

		if secondSlot.IsNewerThan(firstSlot) {
			firstHasSmallerSlot = true
		}
		if firstSlot.IsNewerThan(secondSlot) {
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
