package truetime

import (
	"sync/atomic"
	"testing"
	"time"
)

func TestTrueTimeIntervalBounds(t *testing.T) {
	mockTime := int64(1000)
	uncertainty := int64(10)

	clock := NewClock(uncertainty, func() int64 { return mockTime })
	interval := clock.Now()

	if interval.Earliest != 990 || interval.Latest != 1010 {
		t.Fatalf("expected [990, 1010], got %v", interval)
	}

	stamp := clock.Stamp()
	if stamp != 1010 {
		t.Fatalf("expected Stamp() to be conservative latest bound (1010), got %d", stamp)
	}
}

func TestTrueTimeCommitWaitGuarantee(t *testing.T) {
	// Atomic simulated physical time (in ms)
	var currentTime int64 = 1000

	clock := NewClock(5, func() int64 {
		return atomic.LoadInt64(&currentTime)
	})

	// Transaction commits at t=1000, stamped at latest = 1005
	txStamp := clock.Stamp()

	// Immediately, Earliest is 995, which is NOT > 1005
	if clock.After(txStamp) {
		t.Fatalf("expected clock.After(%d) to be false immediately without waiting", txStamp)
	}

	// Spin a background goroutine that advances time after 10ms
	go func() {
		time.Sleep(10 * time.Millisecond)
		atomic.StoreInt64(&currentTime, 1020) // Earliest becomes 1020 - 5 = 1015 > 1005
	}()

	// CommitWait must block and return only once Earliest > txStamp
	clock.CommitWait(txStamp)

	if !clock.After(txStamp) {
		t.Fatalf("expected clock.After(%d) to be true after commit wait completed", txStamp)
	}
}
