package main

import (
	"strings"
	"testing"
)

func TestDVVHasSeen(t *testing.T) {
	vec := VersionVector{2, 1}

	if !vec.HasSeen(Dot{ReplicaID: 0, Counter: 1}) {
		t.Fatalf("expected vec [2, 1] to have seen (0, 1)")
	}
	if !vec.HasSeen(Dot{ReplicaID: 0, Counter: 2}) {
		t.Fatalf("expected vec [2, 1] to have seen (0, 2)")
	}
	if vec.HasSeen(Dot{ReplicaID: 0, Counter: 3}) {
		t.Fatalf("expected vec [2, 1] NOT to have seen (0, 3)")
	}
	if !vec.HasSeen(Dot{ReplicaID: 1, Counter: 1}) {
		t.Fatalf("expected vec [2, 1] to have seen (1, 1)")
	}
	if vec.HasSeen(Dot{ReplicaID: 1, Counter: 2}) {
		t.Fatalf("expected vec [2, 1] NOT to have seen (1, 2)")
	}
}

func TestDVVDetectsConcurrentWriteWithStaleContext(t *testing.T) {
	replica := newReplica(0, 2)

	// Base write
	replica.Write("v1", newVersionVector(2))
	dotV1 := replica.current.Dot

	// Two clients read v1 simultaneously
	_, ctxA := replica.Read()
	_, ctxB := replica.Read()

	// Client A writes v2
	replica.Write("v2", ctxA)
	dotV2 := replica.current.Dot

	// Invariant: dotV2 must be newer than dotV1
	if dotV2.Counter <= dotV1.Counter {
		t.Fatalf("expected dotV2 counter > dotV1 counter")
	}

	// Client B writes v3 using stale ctxB (which only saw v1)
	replica.Write("v3", ctxB)

	// Critical DVV invariant: Client B's known context did NOT see dotV2!
	// This proves to the system that v3 did NOT observe v2, exposing concurrency.
	if replica.current.KnownVersions.HasSeen(dotV2) {
		t.Fatalf("expected Client B's context NOT to have seen dotV2")
	}
}

func TestDVVIdempotentSync(t *testing.T) {
	repA := newReplica(0, 2)
	repB := newReplica(1, 2)

	repA.Write("item", newVersionVector(2))

	// First sync should succeed
	msg1 := repB.Sync(repA)
	if !strings.Contains(msg1, "accepted new value") {
		t.Fatalf("expected repB to accept new value on first sync, got: %s", msg1)
	}

	// Second sync of same write must be idempotent
	msg2 := repB.Sync(repA)
	if !strings.Contains(msg2, "already have dot") {
		t.Fatalf("expected idempotent rejection on second sync, got: %s", msg2)
	}
}
