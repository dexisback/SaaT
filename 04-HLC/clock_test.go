package main

import (
	"testing"
)

func TestHLCPhysicalAdvance(t *testing.T) {
	clock := NewClock()

	// Physical time advances from 1000 to 1050
	ts1 := clock.LocalEvent(1000)
	if ts1.WallTime != 1000 || ts1.Counter != 0 {
		t.Fatalf("expected ts1 to be (1000, 0), got %v", ts1)
	}

	ts2 := clock.LocalEvent(1050)
	if ts2.WallTime != 1050 || ts2.Counter != 0 {
		t.Fatalf("expected ts2 to be (1050, 0), got %v", ts2)
	}

	if !ts1.IsBefore(ts2) {
		t.Fatalf("expected ts1 to be before ts2")
	}
}

func TestHLCCounterIncrementsWhenPhysicalStalls(t *testing.T) {
	clock := NewClock()

	// Multiple events at the exact same physical millisecond
	ts1 := clock.LocalEvent(1000)
	ts2 := clock.LocalEvent(1000)
	ts3 := clock.LocalEvent(1000)

	if ts1.Counter != 0 || ts2.Counter != 1 || ts3.Counter != 2 {
		t.Fatalf("expected counter to increment [0, 1, 2], got [%d, %d, %d]",
			ts1.Counter, ts2.Counter, ts3.Counter)
	}
	if ts3.WallTime != 1000 {
		t.Fatalf("expected WallTime to remain 1000, got %d", ts3.WallTime)
	}

	if !ts1.IsBefore(ts2) || !ts2.IsBefore(ts3) {
		t.Fatalf("expected strict ordering across counter increments")
	}
}

func TestHLCSkewResilience(t *testing.T) {
	sender := NewClock()
	receiver := NewClock()

	// Sender is ahead at physical time 1050
	sendTS := sender.SendMessage(1050)

	// Receiver's physical clock is lagging behind at 980
	recvTS := receiver.ReceiveMessage(sendTS, 980)

	// Invariant: receiver's WallTime must catch up to sender's WallTime (1050)
	if recvTS.WallTime < sendTS.WallTime {
		t.Fatalf("causality violation: receiver WallTime (%d) is behind sender (%d)",
			recvTS.WallTime, sendTS.WallTime)
	}
	if !sendTS.IsBefore(recvTS) {
		t.Fatalf("causality violation: sendTS must be before recvTS")
	}
}

func TestHLCTiedWallTimeCounterMerge(t *testing.T) {
	nodeA := NewClock()
	nodeB := NewClock()

	tsA := nodeA.SendMessage(1000) // (1000, 0)
	_ = nodeB.SendMessage(1000)    // (1000, 0)

	// Node B makes another event so its counter is 1
	tsB2 := nodeB.LocalEvent(1000) // (1000, 1)

	// Node A receives tsB2 while node A's physical clock still says 1000
	recvA := nodeA.ReceiveMessage(tsB2, 1000) // should jump counter to max(0, 1) + 1 = 2

	if recvA.WallTime != 1000 || recvA.Counter != 2 {
		t.Fatalf("expected receiver to have (1000, 2), got %v (from incoming %v and local %v)",
			recvA, tsB2, tsA)
	}
}
