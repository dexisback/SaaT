package physical

import (
	"testing"
	"time"
)

func TestPhysicalClockDrift(t *testing.T) {
	skew := 50 * time.Millisecond
	clock := New(skew)

	baseTime := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	readTime := clock.Read(baseTime)

	expected := baseTime.Add(skew)
	if !readTime.Equal(expected) {
		t.Fatalf("expected read time %v, got %v", expected, readTime)
	}
}

func TestPhysicalClockNonMonotonicStep(t *testing.T) {
	clock := New(0)
	t1 := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	t2 := t1.Add(10 * time.Millisecond)

	r1 := clock.Read(t1)

	// Simulate backward NTP jump of -50ms before second read
	clock.Step(-50 * time.Millisecond)
	r2 := clock.Read(t2)

	// Non-monotonicity invariant: even though true time moved forward, physical clock went backwards!
	if !r2.Before(r1) {
		t.Fatalf("expected r2 to be before r1 after negative step jump, got r1=%v, r2=%v", r1, r2)
	}
}
