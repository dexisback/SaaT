package lww

import (
	"testing"
	"time"
)

func TestWallRegisterLWW(t *testing.T) {
	reg := NewWallRegister()
	t0 := time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC)
	t1 := t0.Add(1 * time.Second)
	tOld := t0.Add(-1 * time.Second)

	if !reg.Write("initial", t0) {
		t.Fatalf("expected initial write to succeed")
	}

	if !reg.Write("newer_value", t1) {
		t.Fatalf("expected newer write to overwrite")
	}
	if reg.Read().Value != "newer_value" {
		t.Fatalf("expected 'newer_value', got %s", reg.Read().Value)
	}

	if reg.Write("stale_write", tOld) {
		t.Fatalf("expected stale write with older timestamp to be rejected")
	}
	if reg.Read().Value != "newer_value" {
		t.Fatalf("stale write should not have overwritten register value")
	}
}

func TestLamportRegisterLWW(t *testing.T) {
	reg := NewLamportRegister()

	ts1 := LamportTimestamp{Counter: 1, NodeID: 0}
	ts2 := LamportTimestamp{Counter: 2, NodeID: 0}
	tsTieBreaker := LamportTimestamp{Counter: 2, NodeID: 1}

	if !reg.Write("val1", ts1) {
		t.Fatalf("expected first write to succeed")
	}
	if !reg.Write("val2", ts2) {
		t.Fatalf("expected higher counter write to succeed")
	}
	if reg.Read().Value != "val2" {
		t.Fatalf("expected 'val2', got %s", reg.Read().Value)
	}

	if !reg.Write("val_tie_winner", tsTieBreaker) {
		t.Fatalf("expected tie-breaker with higher nodeID to win")
	}
	if reg.Read().Value != "val_tie_winner" {
		t.Fatalf("expected tie breaker winner to be stored")
	}

	if reg.Write("stale", ts1) {
		t.Fatalf("expected lower counter to be rejected")
	}
}
