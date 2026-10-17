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

// Relation represents the true causal relationship between two events
type Relation int

const (
	Same Relation = iota
	HappensBefore
	HappensAfter
	Concurrent
)

func (r Relation) String() string {
	switch r {
	case HappensBefore:
		return "before"
	case HappensAfter:
		return "after"
	case Concurrent:
		return "concurrent"
	default:
		return "same"
	}
}

// Relate determines the ground truth causal relationship from event a to event b.
func (c *Causality) Relate(a, b Event) Relation {
	if a.NodeID == b.NodeID && a.Sequence == b.Sequence {
		return Same
	}
	if c.Before(a, b) {
		return HappensBefore
	}
	if c.Before(b, a) {
		return HappensAfter
	}
	return Concurrent
}

// ConcurrentPairs returns all unordered pairs {a, b} of distinct events that are concurrent.
func (c *Causality) ConcurrentPairs(events []Event) [][2]Event {
	var pairs [][2]Event
	for i := 0; i < len(events); i++ {
		for j := i + 1; j < len(events); j++ {
			if c.Concurrent(events[i], events[j]) {
				pairs = append(pairs, [2]Event{events[i], events[j]})
			}
		}
	}
	return pairs
}

// CausalPairs returns all directed pairs (a, b) where a happened before b.
func (c *Causality) CausalPairs(events []Event) [][2]Event {
	var pairs [][2]Event
	for i := 0; i < len(events); i++ {
		for j := 0; j < len(events); j++ {
			if i != j && c.Before(events[i], events[j]) {
				pairs = append(pairs, [2]Event{events[i], events[j]})
			}
		}
	}
	return pairs
}
