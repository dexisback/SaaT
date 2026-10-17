package matrix

import (
	"testing"
)

func TestMatrixClockLocalAndSend(t *testing.T) {
	node0 := NewClock(0, 3)

	m1 := node0.LocalEvent()
	if m1[0][0] != 1 {
		t.Fatalf("expected node 0's own counter to be 1, got %d", m1[0][0])
	}

	m2 := node0.SendMessage()
	if m2[0][0] != 2 {
		t.Fatalf("expected node 0's send to advance counter to 2, got %d", m2[0][0])
	}
}

func TestMatrixClockKnowledgePropagationAndGC(t *testing.T) {
	node0 := NewClock(0, 3)
	node1 := NewClock(1, 3)
	node2 := NewClock(2, 3)

	// Node 0 produces event 1
	msg0 := node0.SendMessage() // counter = 1 for node 0

	// At this point, node 0 has NOT heard whether node 1 or node 2 saw event 1
	if node0.CanDelete(0, 1) {
		t.Fatalf("node 0 should NOT be able to GC event 1 yet; other nodes have not confirmed receipt")
	}

	// Node 1 receives from Node 0
	msg1 := node1.ReceiveMessage(msg0)

	// Node 2 receives from Node 1 (transitive knowledge)
	node2.ReceiveMessage(msg1)

	// Node 1 sends back to Node 0 confirming it saw event 1
	ack1 := node1.SendMessage()
	node0.ReceiveMessage(ack1)

	// Node 0 knows node 1 saw it, but still doesn't know about node 2
	if node0.CanDelete(0, 1) {
		t.Fatalf("node 0 should still NOT GC; node 2's confirmation hasn't reached node 0")
	}

	// Node 2 sends to Node 0 confirming it also saw event 1
	ack2 := node2.SendMessage()
	node0.ReceiveMessage(ack2)

	// Invariant: NOW every row in node 0's matrix has node 0's counter >= 1!
	if !node0.CanDelete(0, 1) {
		t.Fatalf("expected node 0 to safely garbage collect event 1 now that all nodes confirmed receipt")
	}
}
