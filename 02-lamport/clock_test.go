package lamport

import "testing"

func TestTickIncrements(t *testing.T) {
	c := New(0)
	if got := c.Tick().Counter; got != 1 {
		t.Fatalf("first tick: want 1, got %d", got)
	}
	if got := c.Tick().Counter; got != 2 {
		t.Fatalf("second tick: want 2, got %d", got)
	}
}

func TestReceiveJumpsAhead(t *testing.T) {
	c := New(1)
	got := c.Receive(Timestamp{Counter: 10, NodeID: 0}).Counter
	if got != 11 {
		t.Fatalf("want 11, got %d", got)
	}
}

func TestReceiveNeverGoesBackward(t *testing.T) {
	c := New(1)
	c.Receive(Timestamp{Counter: 10, NodeID: 0}) // now at 11
	got := c.Receive(Timestamp{Counter: 3, NodeID: 0}).Counter
	if got != 12 {
		t.Fatalf("want 12, got %d", got)
	}
}

func TestTieBreakByNodeID(t *testing.T) {
	a := Timestamp{Counter: 5, NodeID: 0}
	b := Timestamp{Counter: 5, NodeID: 2}
	if !a.Less(b) || b.Less(a) {
		t.Fatal("same counter: lower NodeID should be ordered first")
	}
}
