package lamport

import (
	"sort"

	"github.com/dexisback/SaaT/sim"
)

type Stamped struct {
	Event     sim.Event //NodeId, Type,Sequence,MessageID
	TimeStamp Timestamp
}

// stampOne applies the lamport rule for one sweet
// ok=false means a receive whose send hasnt been stamped yet
func stampOne(clock *Clock, event sim.Event, sent map[int]Timestamp) (TimeStamp, bool) {
	switch event.Type {
	case sim.Send:
		timestamp := clock.Send()
		sent[event.MessageID] = timestamp //the stamp travels w the message
		return timestamp, true

	case sim.Receive:
		remote, ok := sent[event.MessageID]
		if !ok {
			return TimeStamp{}, false
		}
		return clock.Receive(remote), true

	default:
		return clock.Tick(), true
	}
}

// replay stamps every event in the trace:
func Replay(events []sim.Event) []Stamped {
	perNode := map[int][]sim.Event{}
	for _, e := range events {
		perNode[e.NodeID] = append(perNode[e.NodeID], e)

	}

	for _, evs := range perNode {
		sort.Slice(evs, func(i int, j int) bool { return evs[i].Sequence < evs[j].Sequence })

	}

	clocks := map[int]*Clock{}
	for node := range perNode {
		clocks[node] = New(node)
	}

	sent := map[int]TimeStamp{}
	next := map[int]int{}
	var out []Stamped

	//same trick as 01-physical : a receive waits until its send is stamped
	for progress := true; progress; {
		progress = false
		for node, evs := range perNode {
			for next[node] < len(evs) {
				e := evs[next[node]]
				ts, ok := stampOne(clocks[node], e, sent)
				if !ok {
					break
				}
				out = append(out, Stamped{Event: e, TimeStamp: ts})
				next[node]++
				progress = true
			}
		}
	}

	sort.Slice(out, func(i int, j int) bool {
		if out[i].Event.NodeID != out[j].Event.NodeID {
			return out[i].Event.NodeID < out[j].Event.NodeID
		}

		return out[i].Event.Sequence < out[j].Event.Sequence
	})
	return out
}
