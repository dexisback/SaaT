package main

import (
	"testing"
)

func TestVectorCompare(t *testing.T) {
	v1 := Vector{1, 0, 0}
	v2 := Vector{2, 0, 0}
	v3 := Vector{0, 1, 0}
	vEqual := Vector{1, 0, 0}

	if res := Compare(v1, v2); res != "before" {
		t.Fatalf("expected v1 before v2, got %s", res)
	}
	if res := Compare(v2, v1); res != "after" {
		t.Fatalf("expected v2 after v1, got %s", res)
	}
	if res := Compare(v1, v3); res != "concurrent" {
		t.Fatalf("expected v1 and v3 to be concurrent, got %s", res)
	}
	if res := Compare(v1, vEqual); res != "equal" {
		t.Fatalf("expected equal vectors, got %s", res)
	}
}

func TestVectorConcurrencyDetection(t *testing.T) {
	node0 := NewClock(0, 3)
	node2 := NewClock(2, 3)

	ev0 := node0.LocalEvent() // [1, 0, 0]
	ev2 := node2.LocalEvent() // [0, 0, 1]

	// This is where Vector clocks outperform Lamport clocks!
	rel := Compare(ev0, ev2)
	if rel != "concurrent" {
		t.Fatalf("expected independent operations on node 0 and node 2 to be concurrent, got %s", rel)
	}
}

func TestVectorCausalPropagation(t *testing.T) {
	node0 := NewClock(0, 3)
	node1 := NewClock(1, 3)

	send0 := node0.SendMessage()     // [1, 0, 0]
	recv1 := node1.ReceiveMessage(send0) // [1, 1, 0]

	if rel := Compare(send0, recv1); rel != "before" {
		t.Fatalf("expected send0 to be strictly before recv1, got %s", rel)
	}
	if recv1[0] != 1 || recv1[1] != 1 {
		t.Fatalf("expected recv1 to have merged [1, 1, 0], got %v", recv1)
	}
}

func TestVectorDefensiveCopy(t *testing.T) {
	node := NewClock(0, 2)
	first := node.LocalEvent() // [1, 0]

	node.LocalEvent() // [2, 0]

	// Ensure the slice returned as `first` wasn't mutated in place
	if first[0] != 1 {
		t.Fatalf("defensive copy violated: earlier vector snapshot mutated to %d", first[0])
	}
}
