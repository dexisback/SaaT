package sim

import (
	"math/rand"
)

// message is actually what travels b/w nodes
type Message struct {
	FromKaun  int
	ToKaun    int
	MessageID int
}

//network will delivery message with artificial random delay, simulating real world jitter

type Network struct {
	inboxes map[int]chan Message
	rng     *rand.Rand
}

func NewNetwork(nodeIDs []int, seed int64) *Network {
	n := *&Network{
		inboxes: make(map[int]chan Message),
		rng:     rand.New(rand.NewSource(seed)),
	}
	for _, id := range nodeIDs {
		n.inboxes[id] = make(chan Message, 10)

	}
	return n
}
