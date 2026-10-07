//It takes the raw event logs from your distributed simulation, reconstructs an ideal timeline ("true time"), applies realistic clock flaws (skews and NTP step jumps), and then runs classic
//distributed systems sanity checks to show how physical time can lie to you.

//AssignTrueTimes :. This function builds an ideal timeline using two rules: local order: a node's action must happen sequentially (spaced apart by localStep) . and network causality: network latency simulation basically

package physical

import (
	"fmt"
	"sort"
	"time"

	"github.com/dexisback/SaaT/sim"
)

var base = time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

const (
	localStep      = 10 * time.Millisecond
	networkLatency = 20 * time.Millisecond
)

// eventkey identifies one event: node + us node ka sequence number
type EventKey struct {
	Node     int
	Sequence int
}

// stamped is an event + both versions of time (so true , and the actual: physical)
type Stamped struct {
	Event    sim.Event
	True     time.Time
	Physical time.Time
}

func AssignTrueTime(events []sim.Event) map[EventKey]time.Time {
	perNode := map[int][]sim.Event{}
	for _, e := range events {
		perNode[e.NodeID] = append(perNode[e.NodeID], e)

	}

	for _, evs := range perNode {
		sort.Slice(evs, func(i int, j int) bool { return evs[i].Sequence < evs[j].Sequence })

	}

	times := map[EventKey]time.Time{}
	sendAt := map[int]time.Time{} //messageID -> true time of the send
	next := map[int]int{}         //node -> index of next unassigned event

	//a receive cannot be placed until its send has been sent, so keep sweeping the nodes until nothing more can be assigned

	for progress := true; progress; {
		progress = false
		for node, evs := range perNode {
			for next[node] < len(evs) {
				e := evs[next[node]]
				t := base
				if next[node] > 0 {
					prev := evs[next[node]-1]
					t = times[EventKey{node, prev.Sequence}].Add(localStep)
				}

				if e.Type == sim.Receive {
					s, ok := sendAt[e.MessageID]
					if !ok {
						break
					}
					//else:
					if arrive := s.Add(networkLatency); arrive.After(t) {
						t = arrive
					}

				}

				times[EventKey{node, e.Sequence}] = t
				if e.Type == sim.Send {
					sendAt[e.MessageID] = t
				}
				next[node]++
				progress = true
			}
		}
	}
	return times
}

//stamp runs every event through its node's skewed clock
//offsets : per-node skew. steps: an NTP styled jump applied just before a specific event  happens

func Stamp(events []sim.Event, offsets map[int]time.Duration, steps map[EventKey]time.Duration) []Stamped {
	trueTimes := AssignTrueTime(events)

	clocks := map[int]*Clock{}
	for node, off := range offsets {
		clocks[node] = New(off)
	}

	ordered := make([]sim.Event, len(events))
	copy(ordered, events)
	sort.Slice(ordered, func(i int, j int) bool {
		if ordered[i].NodeID != ordered[j].NodeID {
			return ordered[i].NodeID < ordered[j].NodeID
		}
		return ordered[i].Sequence < ordered[j].Sequence
	})

	out := make([]Stamped, 0, len(ordered))
	for _, e := range ordered {
		c := clocks[e.NodeID]
		if c == nil {
			c = New(0)
			clocks[e.NodeID] = c
		}
		key := EventKey{e.NodeID, e.Sequence}
		if d, ok := steps[key]; ok {
			c.Step(d)
		}
		tt := trueTimes[key]
		out = append(out, Stamped{Event: e, True: tt, Physical: c.Read(tt)})
	}
	return out
}

func ms(t time.Time) int64 { return t.Sub(base).Milliseconds() }

// Report prints the stamps, then the three ways physical time lies.
func Report(stamped []Stamped) {
	fmt.Println("-- stamps (ms) --")
	for _, s := range stamped {
		fmt.Printf("node %d seq %d %-7v true=%4d  physical=%4d\n",
			s.Event.NodeID, s.Event.Sequence, s.Event.Type, ms(s.True), ms(s.Physical))
	}

	// 1. Effect stamped before its cause.
	sends := map[int]Stamped{}
	recvs := map[int]Stamped{}
	var ids []int
	for _, s := range stamped {
		switch s.Event.Type {
		case sim.Send:
			sends[s.Event.MessageID] = s
			ids = append(ids, s.Event.MessageID)
		case sim.Receive:
			recvs[s.Event.MessageID] = s
		}
	}
	sort.Ints(ids)
	fmt.Println("-- causality check --")
	violations := 0
	for _, id := range ids {
		snd, rcv := sends[id], recvs[id]
		if rcv.Physical.Before(snd.Physical) {
			violations++
			fmt.Printf("VIOLATION msg %d: received (node %d) at %dms, but sent (node %d) at %dms\n",
				id, rcv.Event.NodeID, ms(rcv.Physical), snd.Event.NodeID, ms(snd.Physical))
		}
	}
	if violations == 0 {
		fmt.Println("none")
	}

	// 2. Time going backwards on one node.
	fmt.Println("-- monotonicity check --")
	backwards := 0
	for i := 1; i < len(stamped); i++ {
		a, b := stamped[i-1], stamped[i]
		if a.Event.NodeID == b.Event.NodeID && b.Physical.Before(a.Physical) {
			backwards++
			fmt.Printf("BACKWARDS on node %d: seq %d at %dms, then seq %d at %dms\n",
				a.Event.NodeID, a.Event.Sequence, ms(a.Physical), b.Event.Sequence, ms(b.Physical))
		}
	}
	if backwards == 0 {
		fmt.Println("none")
	}

	// 3. Concurrent events that physical time ranks anyway.
	fmt.Println("-- concurrent events (same true time, different nodes) --")
	for i := 0; i < len(stamped); i++ {
		for j := i + 1; j < len(stamped); j++ {
			a, b := stamped[i], stamped[j]
			if a.Event.Type == sim.Local && b.Event.Type == sim.Local &&
				a.Event.NodeID != b.Event.NodeID && a.True.Equal(b.True) {
				first, second := a, b
				if b.Physical.Before(a.Physical) {
					first, second = b, a
				}
				fmt.Printf("true time identical (%dms), but physical says node %d (%dms) came before node %d (%dms)\n",
					ms(a.True), first.Event.NodeID, ms(first.Physical), second.Event.NodeID, ms(second.Physical))
			}
		}
	}
}
