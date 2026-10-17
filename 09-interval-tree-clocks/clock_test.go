package main

import (
	"testing"
)

func TestITCForkJoinIdentityConservation(t *testing.T) {
	root := newStamp() // owns [0, 1] -> idLeaf(1)

	// Fork into two nodes
	child1, child2 := root.Fork()

	if child1.id.isLeaf && child1.id.value == 1 {
		t.Fatalf("forked child should not own the entire identity space")
	}

	// Join them back
	recombined := Join(child1, child2)

	// Invariant: union of split halves must cover the full space [0, 1]
	normID := normaliseID(recombined.id)
	if !normID.isLeaf || normID.value != 1 {
		t.Fatalf("identity conservation violated: joined id tree normalized to %+v instead of idLeaf(1)", normID)
	}
}

func TestITCEventGrowth(t *testing.T) {
	node := newStamp()

	initialMax := maxValue(node.event)
	node.Event()
	afterFirst := maxValue(node.event)

	if afterFirst <= initialMax {
		t.Fatalf("expected maxValue(event) to grow after Event(), got %d <= %d", afterFirst, initialMax)
	}

	node.Event()
	afterSecond := maxValue(node.event)
	if afterSecond <= afterFirst {
		t.Fatalf("expected maxValue(event) to grow monotonically on subsequent Event(), got %d <= %d",
			afterSecond, afterFirst)
	}
}

func TestITCCausalityAcrossForkAndJoin(t *testing.T) {
	nodeA := newStamp()

	// Node A does something before forking
	nodeA.Event()
	snapshotA := &Stamp{id: copyID(nodeA.id), event: copyEvent(nodeA.event)}

	// Node A forks into Node A and Node B
	nodeA, nodeB := nodeA.Fork()

	// Node A does further work
	nodeA.Event()

	// Invariant: snapshotA happened before current nodeA
	if !snapshotA.HappenedBefore(nodeA) {
		t.Fatalf("expected snapshotA to have happened before current nodeA")
	}

	// Node B does independent work
	nodeB.Event()

	// Both nodes merge their histories back
	merged := Join(nodeA, nodeB)

	if !snapshotA.HappenedBefore(merged) {
		t.Fatalf("expected pre-fork snapshot to have happened before joined history")
	}
}
