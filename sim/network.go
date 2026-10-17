package sim

import (
	"math/rand"
	"sync"
	"time"
)

// message is actually what travels b/w nodes
type Message struct {
	FromKaun  int
	ToKaun    int
	MessageID int
}

//network will delivery message with artificial random delay, simulating real world jitter

type Network struct {
	mu      sync.Mutex // protects rng
	inboxes map[int]chan Message
	rng     *rand.Rand // rng = random number generator
	wg      sync.WaitGroup
}

func NewNetwork(nodeIDs []int, seed int64) *Network {
	n := &Network{
		inboxes: make(map[int]chan Message),
		rng:     rand.New(rand.NewSource(seed)), // simulation produces the exact same sequence everytime you run ts
	}
	for _, id := range nodeIDs {
		n.inboxes[id] = make(chan Message, 50) // buffered channel
	}
	return n
}

// Send pushes a message onto the msg pipeline with random delay (5-50ms).
func (n *Network) Send(from, to, MessageID int) {
	n.mu.Lock()
	delay := time.Duration(5+n.rng.Intn(45)) * time.Millisecond // simulating jitter
	n.mu.Unlock()

	n.wg.Add(1)
	go func() {
		defer n.wg.Done()
		time.Sleep(delay)
		n.inboxes[to] <- Message{FromKaun: from, ToKaun: to, MessageID: MessageID}
	}()
}

// Wait blocks until all in-flight message delivery delays have elapsed and delivered to inboxes
func (n *Network) Wait() {
	n.wg.Wait()
}

// Inbox returns the channel a node should listen on:
func (n *Network) Inbox(nodeID int) chan Message {
	return n.inboxes[nodeID]
}
