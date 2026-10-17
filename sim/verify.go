package sim

import (
	"fmt"
	"strings"
)

// ClockRelation indicates how a clock orders two events
type ClockRelation int

const (
	ClockSame ClockRelation = iota
	ClockBefore
	ClockAfter
	ClockConcurrent
)

func (cr ClockRelation) String() string {
	switch cr {
	case ClockBefore:
		return "before"
	case ClockAfter:
		return "after"
	case ClockConcurrent:
		return "concurrent"
	default:
		return "same"
	}
}

// ClockAdapter is the generic interface that any distributed clock implements to be simulated and verified.
type ClockAdapter interface {
	Name() string
	OnLocal(nodeID int, seq int)
	OnSend(fromNode int, toNode int, seq int, msgID int)
	OnReceive(toNode int, fromNode int, seq int, msgID int)
	Compare(a Event, b Event) ClockRelation
}

// VerificationResult stores the comparison between clock claims and ground truth DAG reachability
type VerificationResult struct {
	ClockName          string
	TotalPairs         int
	CausalPairs        int
	ConcurrentPairs    int
	CausalInversions   int // Ground truth is Before/After, but clock says opposite (FATAL)
	FalseDependencies  int // Ground truth is Concurrent, but clock imposed an order (e.g. Lamport/HLC)
	FalseConcurrency   int // Ground truth is Causal, but clock reported Concurrent
	ExactMatches       int // Ground truth and clock agree perfectly
	Violations         []string
}

func (vr VerificationResult) CausalAccuracy() float64 {
	if vr.CausalPairs == 0 {
		return 100.0
	}
	correct := vr.CausalPairs - vr.CausalInversions - vr.FalseConcurrency
	return float64(correct) / float64(vr.CausalPairs) * 100.0
}

func (vr VerificationResult) ConcurrencyAccuracy() float64 {
	if vr.ConcurrentPairs == 0 {
		return 100.0
	}
	correct := vr.ConcurrentPairs - vr.FalseDependencies
	return float64(correct) / float64(vr.ConcurrentPairs) * 100.0
}

func (vr VerificationResult) Summary() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("=== Verification Report: %s ===\n", vr.ClockName))
	sb.WriteString(fmt.Sprintf("Total Evaluated Event Pairs : %d\n", vr.TotalPairs))
	sb.WriteString(fmt.Sprintf("Ground Truth Causal Pairs   : %d\n", vr.CausalPairs))
	sb.WriteString(fmt.Sprintf("Ground Truth Concurrent     : %d\n", vr.ConcurrentPairs))
	sb.WriteString(fmt.Sprintf("Causal Inversions (FATAL)   : %d\n", vr.CausalInversions))
	sb.WriteString(fmt.Sprintf("False Dependencies          : %d\n", vr.FalseDependencies))
	sb.WriteString(fmt.Sprintf("False Concurrency           : %d\n", vr.FalseConcurrency))
	sb.WriteString(fmt.Sprintf("Causal Invariant Accuracy   : %.1f%%\n", vr.CausalAccuracy()))
	sb.WriteString(fmt.Sprintf("Concurrency Detection Acc   : %.1f%%\n", vr.ConcurrencyAccuracy()))
	if len(vr.Violations) > 0 {
		sb.WriteString("Violations (sample up to 5):\n")
		limit := 5
		if len(vr.Violations) < limit {
			limit = len(vr.Violations)
		}
		for i := 0; i < limit; i++ {
			sb.WriteString(fmt.Sprintf("  - %s\n", vr.Violations[i]))
		}
	}
	return sb.String()
}

// VerifyClock runs the event trace through the adapter, then checks every event pair against ground truth causality.
func VerifyClock(adapter ClockAdapter, events []Event, c *Causality) VerificationResult {
	// 1. Build lookup for send events so receive knows where the message originated
	sends := map[int]Event{}
	for _, e := range events {
		if e.Type == Send {
			sends[e.MessageID] = e
		}
	}

	// 2. Feed events to the clock adapter in chronological order
	for _, e := range events {
		switch e.Type {
		case Local:
			adapter.OnLocal(e.NodeID, e.Sequence)
		case Send:
			adapter.OnSend(e.NodeID, -1, e.Sequence, e.MessageID)
		case Receive:
			fromNode := -1
			if s, ok := sends[e.MessageID]; ok {
				fromNode = s.NodeID
			}
			adapter.OnReceive(e.NodeID, fromNode, e.Sequence, e.MessageID)
		}
	}

	// 3. Verify all pairs
	res := VerificationResult{
		ClockName: adapter.Name(),
	}

	n := len(events)
	for i := 0; i < n; i++ {
		for j := i + 1; j < n; j++ {
			e1, e2 := events[i], events[j]
			res.TotalPairs++

			gt := c.Relate(e1, e2)
			ck := adapter.Compare(e1, e2)

			switch gt {
			case HappensBefore:
				res.CausalPairs++
				if ck == ClockBefore {
					res.ExactMatches++
				} else if ck == ClockAfter {
					res.CausalInversions++
					res.Violations = append(res.Violations,
						fmt.Sprintf("Inversion: %s happened-before %s, but clock says AFTER", e1, e2))
				} else if ck == ClockConcurrent {
					res.FalseConcurrency++
					res.Violations = append(res.Violations,
						fmt.Sprintf("False Concurrency: %s happened-before %s, but clock says CONCURRENT", e1, e2))
				} else {
					res.CausalInversions++
				}

			case HappensAfter:
				res.CausalPairs++
				if ck == ClockAfter {
					res.ExactMatches++
				} else if ck == ClockBefore {
					res.CausalInversions++
					res.Violations = append(res.Violations,
						fmt.Sprintf("Inversion: %s happened-after %s, but clock says BEFORE", e1, e2))
				} else if ck == ClockConcurrent {
					res.FalseConcurrency++
					res.Violations = append(res.Violations,
						fmt.Sprintf("False Concurrency: %s happened-after %s, but clock says CONCURRENT", e1, e2))
				} else {
					res.CausalInversions++
				}

			case Concurrent:
				res.ConcurrentPairs++
				if ck == ClockConcurrent {
					res.ExactMatches++
				} else {
					res.FalseDependencies++
					res.Violations = append(res.Violations,
						fmt.Sprintf("False Dependency: %s and %s are concurrent, but clock says %s", e1, e2, ck))
				}
			}
		}
	}

	return res
}
