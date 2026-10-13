package sim

type EventKey struct {
	Node     int
	Sequence int
}

func KeyOf(event Event) EventKey { return EventKey{event.NodeID, event.Sequence} }

// causality is the ground truth: it answers "did a happen before b?" using only the trace, no clocks involved.
type Causality struct {
	reach map[EventKey]map[EventKey]bool
}

// now function banao isse:
func NewCausality(events []Event) *Causality {

	//what are we doing w sends? in a flag of events (events []Event), you have entires like: Node 0:`Send` Message 100, `Receive` Message 100.
	//this loop filters out only the events where e.Type == Send. and it indexes them by MessageID. (in NewCasuality, when the code encounters `Receive` event, it does: s, ok := sends[e.MessageID])

	sends := map[int]EventKey{} //[{no. : {node, node_sequence}}]
	for _, e := range events {
		if e.Type == Send {
			sends[e.MessageID] = KeyOf(e)
		}
	}
	//below we build the direct edges of a directed graph (causality graph) using an adjancency list (next):
	next := map[EventKey][]EventKey{}
	for _, e := range events {
		k := KeyOf(e) //why not seedha key, value := range events (because its a slice nika)
		if e.Sequence > 0 {
			p := EventKey{e.NodeID, e.Sequence - 1}
			next[p] = append(next[p], k)
		}

		if e.Type == Receive {
			if s, ok := sends[e.MessageID]; ok { //send msg , and append s
				next[s] = append(next[s], k)
			}
		}
	}

	//everything reachable from each event = everything it caused
	c := &Causality{reach: map[EventKey]map[EventKey]bool{}} //// returns true if EventA caused/preceded EventB
	//If c.reach[A][B] is true, it means there is a direct or indirect causal path from $A$ to $B$ in your distributed system.
	for _, e := range events {
		start := KeyOf(e)
		seen := map[EventKey]bool{}
		stack := []EventKey{start}
		for len(stack) > 0 {
			current := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			for _, nx := range next[current] {
				if !seen[nx] {
					seen[nx] = true
					stack = append(stack, nx)

				}
			}
		}
		c.reach[start] = seen
	}
	return c
}

func (c *Causality) Before(a Event, b Event) bool {
	return c.reach[KeyOf(a)][KeyOf(b)]
}

func (c *Causality) Concurrent(a Event, b Event) bool {
	return KeyOf(a) != KeyOf(b) && !c.Before(a, b) && !c.Before(b, a)
}

//{Node: 0, Seq: 1}: {
//        {Node: 1, Seq: 0}: true,
//        {Node: 1, Seq: 1}: true,
//    },
