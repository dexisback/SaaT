package lamport

import (
	"fmt"
	"testing"

	"github.com/dexisback/SaaT/sim"
)

func TestLamportDemo(t *testing.T) {
	for _, s := range Replay(sim.RunBasicScenario(69)) {
		fmt.Printf("node %d sequence %d %v  (counter=%d, node=%d)\n", s.Event.NodeID, s.Event.Sequence, s.TimeStamp, s.TimeStamp.Counter, s.TimeStamp.NodeID)
	}
}
