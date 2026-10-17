package sim

import "fmt"

type EventType int

const (
	Local EventType = iota
	Send
	Receive
)

func (t EventType) String() string {
	switch t {
	case Local:
		return "Local"
	case Send:
		return "Send"
	case Receive:
		return "Receive"
	default:
		return "Unknown"
	}
}

type Event struct {
	NodeID    int
	Type      EventType
	Sequence  int // node's local sequence number
	MessageID int // only meant for sending and receiving messages
}

func (e Event) Key() EventKey {
	return EventKey{Node: e.NodeID, Sequence: e.Sequence}
}

func (e Event) String() string {
	switch e.Type {
	case Local:
		return fmt.Sprintf("Node %d [Local] seq=%d", e.NodeID, e.Sequence)
	case Send:
		return fmt.Sprintf("Node %d [Send] seq=%d msgID=%d", e.NodeID, e.Sequence, e.MessageID)
	case Receive:
		return fmt.Sprintf("Node %d [Receive] seq=%d msgID=%d", e.NodeID, e.Sequence, e.MessageID)
	default:
		return fmt.Sprintf("Node %d [%s] seq=%d", e.NodeID, e.Type, e.Sequence)
	}
}
