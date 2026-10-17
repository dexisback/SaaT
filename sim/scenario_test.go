package sim

import (
	"testing"
)

func TestBasicScenarioCausality(t *testing.T) {
	events := RunBasicScenario(42)
	c := NewCausality(events)

	// Map events by Key for easy querying
	byKey := map[EventKey]Event{}
	for _, e := range events {
		byKey[e.Key()] = e
	}

	n0_local := byKey[EventKey{Node: 0, Sequence: 0}]
	n0_send := byKey[EventKey{Node: 0, Sequence: 1}]
	n1_recv := byKey[EventKey{Node: 1, Sequence: 0}]
	n1_send := byKey[EventKey{Node: 1, Sequence: 1}]
	n2_local := byKey[EventKey{Node: 2, Sequence: 0}]
	n2_recv := byKey[EventKey{Node: 2, Sequence: 1}]

	// Invariant 1: Local sequence ordering
	if !c.Before(n0_local, n0_send) {
		t.Fatalf("expected n0_local to happen before n0_send")
	}

	// Invariant 2: Network message causal edge
	if !c.Before(n0_send, n1_recv) {
		t.Fatalf("expected n0_send to happen before n1_recv")
	}

	// Invariant 3: Transitivity: n0_local -> n1_recv
	if !c.Before(n0_local, n1_recv) {
		t.Fatalf("expected n0_local to transitively happen before n1_recv")
	}
	if !c.Before(n1_send, n2_recv) {
		t.Fatalf("expected n1_send to happen before n2_recv")
	}
	if !c.Before(n0_local, n2_recv) {
		t.Fatalf("expected n0_local to transitively happen before n2_recv")
	}

	// Invariant 4: True concurrency
	// Node 2's local event occurred independently of Node 0's actions
	if !c.Concurrent(n2_local, n0_local) {
		t.Fatalf("expected n2_local and n0_local to be concurrent")
	}
	if !c.Concurrent(n2_local, n0_send) {
		t.Fatalf("expected n2_local and n0_send to be concurrent")
	}

	// But n2_recv is caused by n1_send, so n2_local happens before n2_recv
	if !c.Before(n2_local, n2_recv) {
		t.Fatalf("expected n2_local to happen before n2_recv")
	}
}

func TestCausalChain(t *testing.T) {
	events := RunCausalChain(42)
	c := NewCausality(events)

	byKey := map[EventKey]Event{}
	for _, e := range events {
		byKey[e.Key()] = e
	}

	first := byKey[EventKey{Node: 0, Sequence: 0}]
	last := byKey[EventKey{Node: 3, Sequence: 1}]

	if !c.Before(first, last) {
		t.Fatalf("expected relay chain root to happen before leaf: %s -> %s", first, last)
	}

	// Acyclicity: last cannot happen before first
	if c.Before(last, first) {
		t.Fatalf("causal cycle detected: last happened before first")
	}
}

func TestConcurrentRace(t *testing.T) {
	events := RunConcurrentRace(42)
	c := NewCausality(events)

	byKey := map[EventKey]Event{}
	for _, e := range events {
		byKey[e.Key()] = e
	}

	n0_ev0 := byKey[EventKey{Node: 0, Sequence: 0}]
	n2_ev0 := byKey[EventKey{Node: 2, Sequence: 0}]

	if !c.Concurrent(n0_ev0, n2_ev0) {
		t.Fatalf("expected Node 0 and Node 2 isolated operations to be concurrent")
	}

	pairs := c.ConcurrentPairs(events)
	if len(pairs) == 0 {
		t.Fatalf("expected at least some concurrent pairs in concurrent race scenario")
	}
}

func TestPartitionAndHeal(t *testing.T) {
	events := RunPartitionAndHeal(42)
	c := NewCausality(events)

	byKey := map[EventKey]Event{}
	for _, e := range events {
		byKey[e.Key()] = e
	}

	n0_isolated := byKey[EventKey{Node: 0, Sequence: 0}]
	n1_isolated := byKey[EventKey{Node: 1, Sequence: 0}]

	if !c.Concurrent(n0_isolated, n1_isolated) {
		t.Fatalf("expected partitioned writes to be concurrent")
	}

	// After healing, sync receives follow earlier sends
	n0_sync_send := byKey[EventKey{Node: 0, Sequence: 2}]
	n1_sync_recv := byKey[EventKey{Node: 1, Sequence: 2}]

	if !c.Before(n0_sync_send, n1_sync_recv) {
		t.Fatalf("expected sync message to establish causal edge after healing")
	}
}

// testClockAdapter implements ClockAdapter using ground truth to test the verification engine
type groundTruthAdapter struct {
	c *Causality
}

func (g *groundTruthAdapter) Name() string                                         { return "GroundTruthMock" }
func (g *groundTruthAdapter) OnLocal(nodeID int, seq int)                          {}
func (g *groundTruthAdapter) OnSend(fromNode int, toNode int, seq int, msgID int) {}
func (g *groundTruthAdapter) OnReceive(toNode int, fromNode int, seq int, msgID int) {
}
func (g *groundTruthAdapter) Compare(a Event, b Event) ClockRelation {
	rel := g.c.Relate(a, b)
	switch rel {
	case HappensBefore:
		return ClockBefore
	case HappensAfter:
		return ClockAfter
	case Concurrent:
		return ClockConcurrent
	default:
		return ClockSame
	}
}

func TestVerifyClockEngine(t *testing.T) {
	events := RunBasicScenario(42)
	c := NewCausality(events)

	adapter := &groundTruthAdapter{c: c}
	result := VerifyClock(adapter, events, c)

	if result.CausalInversions != 0 {
		t.Fatalf("expected 0 causal inversions with ground truth adapter, got %d", result.CausalInversions)
	}
	if result.FalseDependencies != 0 {
		t.Fatalf("expected 0 false dependencies, got %d", result.FalseDependencies)
	}
	if result.CausalAccuracy() != 100.0 {
		t.Fatalf("expected 100%% causal accuracy, got %.2f%%", result.CausalAccuracy())
	}
	if result.ConcurrencyAccuracy() != 100.0 {
		t.Fatalf("expected 100%% concurrency accuracy, got %.2f%%", result.ConcurrencyAccuracy())
	}

	summary := result.Summary()
	if len(summary) == 0 {
		t.Fatalf("expected non-empty summary string")
	}
}
