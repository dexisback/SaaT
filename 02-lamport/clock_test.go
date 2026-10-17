package lamport

import (
	"sync"
	"testing"
)

func TestLamportMonotonicity(t *testing.T) {
	clock := NewClock(0)
	t1 := clock.LocalEvent()
	t2 := clock.LocalEvent()

	if t1.Counter >= t2.Counter {
		t.Fatalf("expected t1.Counter < t2.Counter, got t1=%d, t2=%d", t1.Counter, t2.Counter)
	}
	if !t1.isBefore(t2) {
		t.Fatalf("expected t1 to be before t2")
	}
}

func TestLamportCausalityPropagation(t *testing.T) {
	node0 := NewClock(0)
	node1 := NewClock(1)

	// Node 0 advances to 5
	for i := 0; i < 5; i++ {
		node0.LocalEvent()
	}
	sendMsg := node0.SendMessage() // counter = 6

	// Node 1 was at 0
	recvMsg := node1.ReceiveMessage(sendMsg) // counter should be max(0, 6) + 1 = 7

	if recvMsg.Counter <= sendMsg.Counter {
		t.Fatalf("causality violation: receive counter (%d) must be strictly greater than send counter (%d)",
			recvMsg.Counter, sendMsg.Counter)
	}
	if !sendMsg.isBefore(recvMsg) {
		t.Fatalf("expected sendMsg to be ordered before recvMsg")
	}
}

func TestLamportTieBreaking(t *testing.T) {
	// Two timestamps with the same counter
	ts0 := TimeStamp{Counter: 10, NodeID: 0}
	ts1 := TimeStamp{Counter: 10, NodeID: 1}

	if !ts0.isBefore(ts1) {
		t.Fatalf("expected node 0 to break tie before node 1 when counters are equal")
	}
	if ts1.isBefore(ts0) {
		t.Fatalf("expected node 1 NOT to be before node 0")
	}
}

func TestLamportThreadSafety(t *testing.T) {
	clock := NewClock(0)
	var wg sync.WaitGroup
	iterations := 1000

	wg.Add(2)
	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			clock.LocalEvent()
		}
	}()
	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			clock.LocalEvent()
		}
	}()
	wg.Wait()

	if clock.Counter != iterations*2 {
		t.Fatalf("expected final counter to be %d, got %d", iterations*2, clock.Counter)
	}
}
