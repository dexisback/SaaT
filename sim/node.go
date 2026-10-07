package sim

//a node needs to :
// - keep a local sequence counter
// - perform actions (local/send/receive)
// - record every `event` it produces into a shared log so i can trace the thing afterwards
import "sync"

// collects events from all nodes, in the order they actually happened in wall-clock time (not logical time)
type EventLog struct {
	mu     sync.Mutex //the lock / the key
	events []Event    // the shared data resource
}

func (l *EventLog) Record(e Event) {
	l.mu.Lock()
	defer l.mu.Unlock()
	out := make([]Event, len(l.events)) //buffer of length of events, e)
	copy(out, l.events)
	l.events = out
	l.events = append(l.events, e)
}

func (l *EventLog) All() []Event {
	l.mu.Lock()
	defer l.mu.Unlock()
	out := make([]Event, len(l.events))
	copy(out, l.events)
	return out
}

// node walks through actions and emits events as it goes
type Node struct {
	ID  int
	seq int //seq acts as local sequence counter for an individual node
	net *Network
	log *EventLog
}

func NewNode(id int, net *Network, log *EventLog) *Node {
	return &Node{ID: id, net: net, log: log}

}

func (n *Node) nextSeq() int {
	//i could have written it like:
	// val := n.seq
	// n.seq++
	// return val
	//go does NOT allow something like return n.seq++

	n.seq++
	return n.seq - 1
}

// DoLocal records a plain local event:
func (n *Node) DoLocal() {
	n.log.Record(Event{NodeID: n.ID, Type: Local, Sequence: n.nextSeq()})
}

// DoReceive sends a message to another node and records the Send event:
func (n *Node) DoSend(to, msgID int) {
	n.log.Record(Event{NodeID: n.ID, Type: Send, Sequence: n.nextSeq(), MessageID: msgID}) //records the event
	n.net.Send(n.ID, to, msgID)                                                            //and then puts it on the wire(sending wala wire)
}

func (n *Node) DoReceive() Message {
	msg := <-n.net.Inbox(n.ID) //a node genuinely waits for a message to arrive before it can record receiving it
	n.log.Record(Event{NodeID: n.ID, Type: Receive, Sequence: n.nextSeq(), MessageID: msg.MessageID})
	return msg
}

//sequence is per-node, starting at 0. every node only knows its own local ordering, never anyone elses
// EventLog is shared and mutex-protected
// When a node calls Record() or All(), it immediately calls l.mu.Lock(). If no other node has the key, it grabs it and proceeds. If another node currently holds the key, the current node pauses (blocks) right there and waits until the key is available.
//
