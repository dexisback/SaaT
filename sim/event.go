//in a distributed systum, every node's history
// //is just a sequence of 3 kinds of things:
// //local event/send event and receive event
// //local mtlb node does something internally (ex: write to its own disk)
// //sending event mtlb sending to another node obv

package sim

type EventType int

const (
	Local EventType = iota
	Send
	Receive
)

type Event struct {
	NodeID    int
	Type      EventType
	Sequence  int //node's local sequence number
	MessageID int //only meant for sending and receiving messages
}

//why sequence and not timestamps? because /sim should doesnt have any business in clock logic. woh clock folders deal karlenge
//
