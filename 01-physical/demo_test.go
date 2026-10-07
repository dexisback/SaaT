package physical

import (
	"fmt"
	"testing"
	"time"

	"github.com/dexisback/SaaT/sim"
)

func TestPhysicalDemo(t *testing.T) {
	events := sim.RunBasicScenario(69)

	offsets := map[int]time.Duration{
		0: 0,
		1: -50 * time.Millisecond, //node 1 runs 50ms behind
		2: 5 * time.Millisecond,   //node 2 runs 5ms ahead

	}

	fmt.Println("===A: skew only === ")
	Report(Stamp(events, offsets, nil))

	fmt.Println("===B: skew + NTP step on node 1 ===")
	steps := map[EventKey]time.Duration{
		{1, 1}: -100 * time.Millisecond, //clock jumps back before node 1's second event
	}
	Report(Stamp(events, offsets, steps))
}
