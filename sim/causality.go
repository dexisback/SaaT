package sim

type EventKey struct {
	Node int
	Seq  int
}

func KeyOf(e Event) EventKey { return EventKey{e.NodeID, e.Sequence} }

// Causality is the ground truth: it answers "did a happen before b?"
// using only the trace, no clocks involved.
type Causality struct {
	reach map[EventKey]map[EventKey]bool
}

func NewCausality(events []Event) *Causality {
	sends := map[int]EventKey{}
	for _, e := range events {
		if e.Type == Send {
			sends[e.MessageID] = KeyOf(e)
		}
	}

	// direct edges: previous event on the same node, and send -> receive
	next := map[EventKey][]EventKey{}
	for _, e := range events {
		k := KeyOf(e)
		if e.Sequence > 0 {
			p := EventKey{e.NodeID, e.Sequence - 1}
			next[p] = append(next[p], k)
		}
		if e.Type == Receive {
			if s, ok := sends[e.MessageID]; ok {
				next[s] = append(next[s], k)
			}
		}
	}

	// everything reachable from each event = everything it caused
	c := &Causality{reach: map[EventKey]map[EventKey]bool{}}
	for _, e := range events {
		start := KeyOf(e)
		seen := map[EventKey]bool{}
		stack := []EventKey{start}
		for len(stack) > 0 {
			cur := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			for _, nx := range next[cur] {
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

func (c *Causality) Before(a, b Event) bool {
	return c.reach[KeyOf(a)][KeyOf(b)]
}

func (c *Causality) Concurrent(a, b Event) bool {
	return KeyOf(a) != KeyOf(b) && !c.Before(a, b) && !c.Before(b, a)
}
