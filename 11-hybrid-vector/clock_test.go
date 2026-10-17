package hybridvector

import (
	"testing"
)

func TestHybridVectorCompare(t *testing.T) {
	v1 := HybridVector{
		{WallTime: 1000, Counter: 0},
		{WallTime: 0, Counter: 0},
	}
	v2 := HybridVector{
		{WallTime: 1000, Counter: 1},
		{WallTime: 0, Counter: 0},
	}
	v3 := HybridVector{
		{WallTime: 0, Counter: 0},
		{WallTime: 1000, Counter: 0},
	}

	if rel := Compare(v1, v2); rel != Before {
		t.Fatalf("expected v1 Before v2 (counter tie-break), got %s", rel)
	}
	if rel := Compare(v2, v1); rel != After {
		t.Fatalf("expected v2 After v1, got %s", rel)
	}
	if rel := Compare(v1, v3); rel != Concurrent {
		t.Fatalf("expected independent hybrid slots to be Concurrent, got %s", rel)
	}
	if rel := Compare(v1, v1); rel != Equal {
		t.Fatalf("expected v1 Equal v1, got %s", rel)
	}
}

func TestHybridVectorCausalityAndConcurrency(t *testing.T) {
	node0 := NewClock(0, 3)
	node1 := NewClock(1, 3)
	node2 := NewClock(2, 3)

	// Node 0 local at 1000, Node 2 local at 1005
	ev0 := node0.LocalEvent(1000)
	ev2 := node2.LocalEvent(1005)

	// Hybrid Vector retains the superpower of Vector clocks: Concurrency Detection!
	if rel := Compare(ev0, ev2); rel != Concurrent {
		t.Fatalf("expected ev0 and ev2 to be Concurrent, got %s", rel)
	}

	// Node 0 sends to Node 1
	send0 := node0.SendMessage(1010)
	recv1 := node1.ReceiveMessage(send0, 1015)

	// Invariant: causality is strictly preserved
	if rel := Compare(send0, recv1); rel != Before {
		t.Fatalf("expected send0 to be Before recv1, got %s", rel)
	}
}
