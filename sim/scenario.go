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
	return log.All()
}
