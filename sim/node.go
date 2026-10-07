package sim

//a node needs to :
// - keep a local sequence counter
// - perform actions (local/send/receive)
// - record every `event` it produces into a shared log so i can trace the thing afterwards
import "sync"

// collects events from all nodes, in the order they actually happened in wall-clock time (not logical time)
type EventLog struct {
	mu     sync.Mutex
	events []Event
}

func (l *EventLog) record(e Event) <-chan []Event {
	l.mu.Lock()
	defer l.mu.Unlock()
	out := make([]Event, len(l.events)) //buffer of length of events, dayum
	copy(out, l.events)
	return out

}
