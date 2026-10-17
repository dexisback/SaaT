package itc

import "fmt"

//every clock (mostly vector clocks) have number of nodes fixed. but then if new nodes keep on incming, every existing clock becomes wrong.
//Similarly, if a node leaves, its slot wastes space forever.

//ITC solves this without any central authority handing out nodeIDs

//how? this analogy from claude helped me visualise it better:
// magine the number line from 0 to 1. At the start, one node owns the whole line.
//  When it wants to add a node, it splits its piece in half and gives one half away. When a node leaves, its piece gets merged back. No coordinator needed, no fixed size, no wasted slots.

//clock: ID + Event . ID = which piece of the identity space this node owns. represented as a binary tree with 0s(doesnt own this piece) or 1s(owns)
//event: how many events have happened in each piece of the space. also a binary tree, but leaves are integers instead of 0/1

// (i hate trees)

//fork: one node becomes two (we split the ID tree, each child gets half)
//event(can be local or send): increment the event tree in the region this node'ID covers
//join: two nodes merge back into one. Union the ID trees, take the max of the event trees

// implementing ID tree (will consist of a leaf(which can be 0 or 1) + a branch
type IDTree struct {
	isLeaf bool
	value  int
	left   *IDTree //only used when isLeaf=false
	right  *IDTree //only used when isLeaf=false
}

func idLeaf(value int) *IDTree {
	return &IDTree{
		isLeaf: true, value: value,
	}
}

func idBranch(left *IDTree, right *IDTree) *IDTree {
	return &IDTree{isLeaf: false, left: left, right: right}
}

// copyID makes a deep copy of an id tree:
func copyID(original *IDTree) *IDTree {
	if original.isLeaf {
		return idLeaf(original.value)
	}
	return idBranch(copyID(original.left), copyID(original.right))

}

// normaliseID simplifies an ID tree: a branch where both children are the same leaf becomes that leaf
func normaliseID(tree *IDTree) *IDTree {
	if tree.isLeaf {
		return tree
	}
	leftNorm := normaliseID(tree.left)
	rightNorm := normaliseID(tree.right)

	if leftNorm.isLeaf && rightNorm.isLeaf && leftNorm.value == rightNorm.value {
		return idLeaf(leftNorm.value)
	}

	return idBranch(leftNorm, rightNorm)
}

// splitId divides an ID tree into two halves, for forking
func splitID(tree *IDTree) (*IDTree, *IDTree) {
	if tree.isLeaf {
		if tree.value == 0 {
			return idLeaf(0), idLeaf(0) //own nothing -- cant split
		}

		//else: own everything, give left half to one, right half to another
		return idBranch(idLeaf(1), idLeaf(0)), idBranch(idLeaf(0), idLeaf(1))
	}

	// If left side is 0, recurse on right
	if tree.left.isLeaf && tree.left.value == 0 {
		r1, r2 := splitID(tree.right)
		return idBranch(idLeaf(0), r1), idBranch(idLeaf(0), r2)
	}

	// If right side is 0, recurse on left
	if tree.right.isLeaf && tree.right.value == 0 {
		l1, l2 := splitID(tree.left)
		return idBranch(l1, idLeaf(0)), idBranch(l2, idLeaf(0))
	}

	// Both sides are owned: give left to one, right to the other
	return idBranch(tree.left, idLeaf(0)), idBranch(idLeaf(0), tree.right)
}

// unionID combined two ID trees: a region is owned if either side own it
func unionID(first *IDTree, second *IDTree) *IDTree {
	if first.isLeaf && first.value == 1 {
		return idLeaf(1)
	}
	if second.isLeaf && second.value == 1 {
		return idLeaf(1)
	}
	if first.isLeaf && first.value == 0 {
		return copyID(second)
	}
	if second.isLeaf && second.value == 0 {
		return copyID(first)
	}
	return normaliseID(idBranch(
		unionID(first.left, second.left),
		unionID(first.right, second.right),
	))
}

func (tree *IDTree) String() string {
	if tree.isLeaf {
		return fmt.Sprintf("%d", tree.value)
	}
	return fmt.Sprintf("(%s, %s)", tree.left.String(), tree.right.String())
}




//-------------------------------------



//now: building the event tree:


type EventTree struct {
	isLeaf bool
	value  int        // base count: added to everything below this node
	left   *EventTree // only used when isLeaf=false
	right  *EventTree // only used when isLeaf=false
}

func eventLeaf(value int) *EventTree {
	return &EventTree{isLeaf: true, value: value}
}

func eventBranch(base int, left *EventTree, right *EventTree) *EventTree {
	return &EventTree{isLeaf: false, value: base, left: left, right: right}
}

// copyEvent makes a deep copy of an event tree.
func copyEvent(original *EventTree) *EventTree {
	if original.isLeaf {
		return eventLeaf(original.value)
	}
	return eventBranch(original.value, copyEvent(original.left), copyEvent(original.right))
}



//minValue returns the minimum counter value anywhere in an event tree:
func minValue(tree *EventTree) int {
	if tree.isLeaf {
		return tree.value
	}
	leftMin := minValue(tree.left)
	rightMin := minValue(tree.right)
	if leftMin < rightMin {
		return tree.value + leftMin
	} //else:
	return tree.value + rightMin
}




//maxValue returns the maximum counter value anywhere in an event tree
func maxValue(tree *EventTree) int {
	if tree.isLeaf {
		return tree.value
	}
	leftMax := maxValue(tree.left)
	rightMax := maxValue(tree.right)
	if leftMax > rightMax {
		return tree.value + leftMax
	}
	return tree.value + rightMax
}


//smol helper:
func subtractFromEvent(tree *EventTree, delta int) *EventTree {
	if tree.isLeaf {
		return eventLeaf(tree.value - delta)
	}
	return eventBranch(tree.value-delta, tree.left, tree.right)
}

//normaliseEvent simplifies an event tree by lifting common base values:

func normalizeEvent(tree *EventTree) *EventTree {
	if tree.isLeaf {
		return tree
	}
	leftNorm := normalizeEvent(tree.left)
	rightNorm := normalizeEvent(tree.right)

	// if both children are equal leaves, collapse to one leaf
	if leftNorm.isLeaf && rightNorm.isLeaf && leftNorm.value == rightNorm.value {
		return eventLeaf(tree.value + leftNorm.value)
	}

	// lift the common minimum up into the base
	leftMin := minValue(leftNorm)
	rightMin := minValue(rightNorm)
	commonMin := leftMin
	if rightMin < commonMin {
		commonMin = rightMin
	}

	if commonMin > 0 {
		return eventBranch(
			tree.value+commonMin,
			subtractFromEvent(leftNorm, commonMin),
			subtractFromEvent(rightNorm, commonMin),
		)
	}
	return eventBranch(tree.value, leftNorm, rightNorm)
}

// fillEvent increments the event tree in the regions covered by an ID tree.
func fillEvent(eventTree *EventTree, idTree *IDTree) *EventTree {
	if idTree.isLeaf && idTree.value == 1 {
		// this ID owns this whole region: set it to the max seen here
		return eventLeaf(maxValue(eventTree))
	}
	if idTree.isLeaf && idTree.value == 0 {
		// this ID owns nothing here: leave it unchanged
		return copyEvent(eventTree)
	}
	if eventTree.isLeaf {
		// need to split the leaf to fill only part of it
		filledLeft := fillEvent(eventLeaf(eventTree.value), idTree.left)
		filledRight := fillEvent(eventLeaf(eventTree.value), idTree.right)
		return normalizeEvent(eventBranch(0, filledLeft, filledRight))
	}
	filledLeft := fillEvent(eventTree.left, idTree.left)
	filledRight := fillEvent(eventTree.right, idTree.right)
	return normalizeEvent(eventBranch(eventTree.value, filledLeft, filledRight))
}



// growEvent increments by 1 in the region an ID tree owns.
func growEvent(eventTree *EventTree, idTree *IDTree) *EventTree {
	if idTree.isLeaf && idTree.value == 1 {
		if eventTree.isLeaf {
			return eventLeaf(eventTree.value + 1)
		}
		return normalizeEvent(eventBranch(eventTree.value+1, eventTree.left, eventTree.right))
	}
	if idTree.isLeaf && idTree.value == 0 {
		return copyEvent(eventTree)
	}
	if eventTree.isLeaf {
		grownLeft := growEvent(eventLeaf(eventTree.value), idTree.left)
		grownRight := growEvent(eventLeaf(eventTree.value), idTree.right)
		return normalizeEvent(eventBranch(0, grownLeft, grownRight))
	}
	grownLeft := growEvent(eventTree.left, idTree.left)
	grownRight := growEvent(eventTree.right, idTree.right)
	return normalizeEvent(eventBranch(eventTree.value, grownLeft, grownRight))
}


// joinEvents takes the slot-wise maximum of two event trees.

func joinEvents(first *EventTree, second *EventTree) *EventTree {
	if first.isLeaf && second.isLeaf {
		if first.value > second.value {
			return eventLeaf(first.value)
		}
		return eventLeaf(second.value)
	}
	if first.isLeaf {
		return joinEvents(
			eventBranch(first.value, eventLeaf(0), eventLeaf(0)),
			second,
		)
	}
	if second.isLeaf {
		return joinEvents(
			first,
			eventBranch(second.value, eventLeaf(0), eventLeaf(0)),
		)
	}
	baseMax := first.value
	if second.value > baseMax {
		baseMax = second.value
	}
	return normalizeEvent(eventBranch(
		baseMax,
		joinEvents(
			subtractFromEvent(first.left, baseMax-first.value),
			subtractFromEvent(second.left, baseMax-second.value),
		),
		joinEvents(
			subtractFromEvent(first.right, baseMax-first.value),
			subtractFromEvent(second.right, baseMax-second.value),
		),
	))
}


func (tree *EventTree) String() string {
	if tree.isLeaf {
		return fmt.Sprintf("%d", tree.value)
	}
	return fmt.Sprintf("(%d, %s, %s)", tree.value, tree.left.String(), tree.right.String())
}


//-------------------------------------------



//stamp:
type Stamp struct {
	id    *IDTree
	event *EventTree
}


func newStamp() *Stamp {
	return &Stamp{
		id:    idLeaf(1),        // start owning the whole identity space
		event: eventLeaf(0),    // no events yet
	}
}


func (stamp *Stamp) String() string {
	return fmt.Sprintf("ID:%s EVENT:%s", stamp.id.String(), stamp.event.String())
}






//------------------------




//implementing fork:

func (stamp *Stamp) Fork() (*Stamp, *Stamp) {
	firstID, secondID := splitID(stamp.id)
	return &Stamp{id: firstID, event: copyEvent(stamp.event)},
		&Stamp{id: secondID, event: copyEvent(stamp.event)}
}


//implementing event: which records a local event on this stamp:
func (stamp *Stamp) Event() {
	filled := fillEvent(stamp.event, stamp.id)
	if maxValue(filled) > maxValue(stamp.event) {
		stamp.event = filled
		return
	}
	stamp.event = growEvent(stamp.event, stamp.id)
}



//join merges two stamps back into one: for whne a node is shutting down or two nodes want to merge their histories
func Join(first *Stamp, second *Stamp) *Stamp {
	return &Stamp{
		id:    unionID(first.id, second.id),
		event: joinEvents(first.event, second.event),
	}
}


func NewStamp() *Stamp {
	return newStamp()
}

func (stamp *Stamp) HappenedBefore(other *Stamp) bool {
	return maxValue(stamp.event) <= minValue(other.event) &&
		maxValue(stamp.event) < maxValue(other.event)
}