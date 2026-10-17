//this reproduces the story we'll replay through every clock, to visualise/etc
// //3 nodes, mix of local events and msg passing, including one case where two nodes do things independently (concurrency)

package sim

import "sync"

func RunBasicScenario(seed int64) []Event {
	nodeIDs := []int{0, 1, 2} //three nodes
	net := NewNetwork(nodeIDs, seed)
	log := &EventLog{}

	nodes := make(map[int]*Node) //this
	for _, id := range nodeIDs {
		nodes[id] = NewNode(id, net, log)
	}

	var wg sync.WaitGroup

	//node 0: local event, then sends message 100 to node 1
	wg.Add(1)
	go func() {
		defer wg.Done()
		n := nodes[0]
		n.DoLocal()      //record the thang
		n.DoSend(1, 100) //100 = msg id, to node 1
	}()
	//node 1:
	wg.Add(1)
	go func() {
		defer wg.Done()
		n := nodes[1]
		n.DoReceive() //gets msg 100
		n.DoSend(2, 101)
	}()

	// node 2: does its OWN local event (unrelated to 0 and 1)
	// then receives msg 101 from node 1
	wg.Add(1)
	go func() {
		defer wg.Done()
		n := nodes[2]
		n.DoLocal()   //this is concurrent w node 0's local event
		n.DoReceive() //gets msg 101
	}()
	wg.Wait()
	net.Wait()
	return log.All()
}

// RunCausalChain runs a strict relay: 0 -> 1 -> 2 -> 3
func RunCausalChain(seed int64) []Event {
	nodeIDs := []int{0, 1, 2, 3}
	net := NewNetwork(nodeIDs, seed)
	log := &EventLog{}

	nodes := make(map[int]*Node)
	for _, id := range nodeIDs {
		nodes[id] = NewNode(id, net, log)
	}

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		nodes[0].DoLocal()
		nodes[0].DoSend(1, 200)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		nodes[1].DoReceive()
		nodes[1].DoLocal()
		nodes[1].DoSend(2, 201)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		nodes[2].DoReceive()
		nodes[2].DoLocal()
		nodes[2].DoSend(3, 202)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		nodes[3].DoReceive()
		nodes[3].DoLocal()
	}()

	wg.Wait()
	net.Wait()
	return log.All()
}

// RunConcurrentRace runs two independent nodes operating concurrently before converging on node 1
func RunConcurrentRace(seed int64) []Event {
	nodeIDs := []int{0, 1, 2}
	net := NewNetwork(nodeIDs, seed)
	log := &EventLog{}

	nodes := make(map[int]*Node)
	for _, id := range nodeIDs {
		nodes[id] = NewNode(id, net, log)
	}

	var wg sync.WaitGroup

	// Node 0 does two local events and then sends to 1
	wg.Add(1)
	go func() {
		defer wg.Done()
		nodes[0].DoLocal()
		nodes[0].DoLocal()
		nodes[0].DoSend(1, 300)
	}()

	// Node 2 does two independent local events and then sends to 1
	wg.Add(1)
	go func() {
		defer wg.Done()
		nodes[2].DoLocal()
		nodes[2].DoLocal()
		nodes[2].DoSend(1, 301)
	}()

	// Node 1 waits to receive from both
	wg.Add(1)
	go func() {
		defer wg.Done()
		nodes[1].DoReceive()
		nodes[1].DoLocal()
		nodes[1].DoReceive()
	}()

	wg.Wait()
	net.Wait()
	return log.All()
}

// RunPartitionAndHeal simulates concurrent isolated operations on two replicas followed by mutual sync
func RunPartitionAndHeal(seed int64) []Event {
	nodeIDs := []int{0, 1}
	net := NewNetwork(nodeIDs, seed)
	log := &EventLog{}

	nodes := make(map[int]*Node)
	for _, id := range nodeIDs {
		nodes[id] = NewNode(id, net, log)
	}

	var wg sync.WaitGroup

	// Phase 1: Isolated local actions
	wg.Add(2)
	go func() {
		defer wg.Done()
		nodes[0].DoLocal()
		nodes[0].DoLocal()
	}()
	go func() {
		defer wg.Done()
		nodes[1].DoLocal()
		nodes[1].DoLocal()
	}()
	wg.Wait()

	// Phase 2: Heal & exchange sync messages
	wg.Add(2)
	go func() {
		defer wg.Done()
		nodes[0].DoSend(1, 400)
		nodes[0].DoReceive() // receive node 1's sync
	}()
	go func() {
		defer wg.Done()
		nodes[1].DoReceive() // receive node 0's sync
		nodes[1].DoSend(0, 401)
	}()
	wg.Wait()
	net.Wait()

	return log.All()
}
