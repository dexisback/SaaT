//basically, a version vector tells you the causal history you know about. 
//but a dotted version vector additionally tells you the specific write that produced a version, so concurrent writes arent accidentally written
//when would the version vector fail? -- suppose you have a replica v1 and two clients that read it and do something something w it (Client A and Client B)
//suppose client A receives it [1,0] and does a write. so v 2becomes [2 0 ]
//client b also reads v1. [1 0 ] and does some write to make it v3. eventually becomes [3 0]
//now you have v2 = [2 0 ] and v3=[3 0 ], but notice the subtlety -- [2, 0 ] and [3,0] suggests a sequential history, even though client B never saw v2. a vector's counter tells us about causal history, but this simplistic scheme doesnt preserve the identity of each individual write.



//what do the dots add? a dotted version vector separates 2 pieces of information: 1 -- Context (the version vector): which prior writes the client knows about
//2 -- Dot(replicaID, counter) : the identity of the particular write being made. 
//now the entire pipeline would become: 
//original value = Context: [1 0] and Dot : (R1, 1)
//client A's write = v2 = Context : [1 0] and Dot : (R1, 2)
//client B's write = v3 = Context: [1 0] and Dot : (R1, 3)
//now the system tracks the new write separately form the context the client supported


//therefore, a replica can now distinguish b/w:
// - a write that has observed an earlier write and causally follows it .
// - a write that was made independently, without observing that earlier write. 

//if v2 and v3 are concurrent, the system can preserve both as conflicting versions instead of silently discarding one. the application code can then resolve the conflict or do wtv it wants to w the info. point is we solved that info not being able to get problem



package main

import "fmt"

//a version vector has one slot per replica, same as before 

type VersionVector []int

func newVersionVector(numberOfReplicas int) VersionVector {
	return make(VersionVector, numberOfReplicas)
}

func copyOf(original VersionVector) VersionVector {
	copied := make(VersionVector, len(original))
	copy(copied, original)
	return copied
}

// HasSeen checks whether a version vector includes a specific dot
func (vector VersionVector) HasSeen(dot Dot) bool {
	if dot.ReplicaID < 0 || dot.ReplicaID >= len(vector) {
		return false
	}
	return vector[dot.ReplicaID] >= dot.Counter
}

//dot identifies one specific write event : which replica wrote it and at what counter value . 
type Dot struct {
	ReplicaID int
	Counter   int
}

func (dot Dot) String() string {
	return fmt.Sprintf("(replica %d, counter %d)", dot.ReplicaID, dot.Counter)
}

//dotted value is what a replica stores : the value itself, the dot that created it, and the version vector of everything the writer had seen 
type DottedValue struct {
	Value         string
	Dot           Dot
	KnownVersions VersionVector //what the client had seen when it wrote
}

//replica stores : key's dotted value + current version vector:
type Replica struct {
	replicaID     int
	current       *DottedValue //nil will mean nothing has been written yet 
	versionVector VersionVector
}

func newReplica(replicaID int, numberOfReplicas int) *Replica {
	return &Replica{
		replicaID:     replicaID,
		versionVector: newVersionVector(numberOfReplicas),
	}
}

//read returns the current value and the current version vector . NOTE: the client must keep the version vector and send it back on write, so the replica knows EXACTLY what the client had seen
func (replica *Replica) Read() (string, VersionVector) {
	if replica.current == nil {
		return "", copyOf(replica.versionVector)
	}
	//else: 
	return replica.current.Value, copyOf(replica.versionVector)
}

//write stores a new value. clientContext is the version vector the client received when it last read. the replica uses it to detect whether this write is on top of the current value or a stale one 
func (replica *Replica) Write(newValue string, clientContext VersionVector) {
	replica.versionVector[replica.replicaID]++
	currentCounter := replica.versionVector[replica.replicaID]

	newDot := Dot{ReplicaID: replica.replicaID, Counter: currentCounter}
	replica.current = &DottedValue{
		Value:         newValue,
		Dot:           newDot,
		KnownVersions: copyOf(clientContext),
	}

	// merge the client's known versions into our vector
	for slot := range replica.versionVector {
		if clientContext[slot] > replica.versionVector[slot] {
			replica.versionVector[slot] = clientContext[slot]
		}
	}
}

//sync merges another replica's dotted value into this one. this is where the dot saves us: we check whether the incoming dot is alr known to us (its counter is inside our version vector ). If it is, we've alr seen that write. if not, its new 
func (replica *Replica) Sync(other *Replica) string {
	if other.current == nil {
		return "other replica has no value, nothing to sync"
	}

	incomingDot := other.current.Dot
	incomingCounter := incomingDot.Counter
	ourKnowledgeOfThatReplica := replica.versionVector[incomingDot.ReplicaID]

	if incomingCounter <= ourKnowledgeOfThatReplica {
		return fmt.Sprintf("already have dot %v, nothing to do", incomingDot)
	}

	// It's a new write we haven't seen: accept the value
	replica.current = &DottedValue{
		Value:         other.current.Value,
		Dot:           other.current.Dot,
		KnownVersions: copyOf(other.current.KnownVersions),
	}

	// Merge version vectors so we record our new causal knowledge
	for slot := range replica.versionVector {
		if other.versionVector[slot] > replica.versionVector[slot] {
			replica.versionVector[slot] = other.versionVector[slot]
		}
	}

	return fmt.Sprintf("accepted new value '%s' with dot %v", other.current.Value, incomingDot)
}

func main() {
	fmt.Println("=== the scenario that breaks plain version vectors ===")
	fmt.Println()

	replicaR := newReplica(0, 2) // the one replica storing the key

	// initial write to get things started
	replicaR.Write("v1", newVersionVector(2))
	fmt.Println("replica writes v1, vector:", replicaR.versionVector)

	// client 1 and client 2 both read the same version
	_, contextSeenByClient1 := replicaR.Read()
	_, contextSeenByClient2 := replicaR.Read()
	fmt.Println("client 1 reads, sees context:", contextSeenByClient1)
	fmt.Println("client 2 reads, sees context:", contextSeenByClient2)

	// client 1 writes v2 on top of what it saw
	replicaR.Write("v2", contextSeenByClient1)
	dotV2 := replicaR.current.Dot
	fmt.Println()
	fmt.Println("client 1 writes v2, vector:", replicaR.versionVector)
	fmt.Println("current dot:", replicaR.current.Dot)

	// client 2 writes v3 on top of what IT saw (the old context)
	// with plain version vectors this would silently overwrite v2
	// with dotted version vectors, the replica knows client 2 hadn't seen v2
	replicaR.Write("v3", contextSeenByClient2)
	fmt.Println()
	fmt.Println("client 2 writes v3 (with stale context), vector:", replicaR.versionVector)
	fmt.Println("current dot:", replicaR.current.Dot)
	fmt.Println("current value:", replicaR.current.Value)
	fmt.Println("client 2's known versions when it wrote:", replicaR.current.KnownVersions)
	fmt.Println("did client 2 see v2 before writing v3?", replicaR.current.KnownVersions.HasSeen(dotV2))

	fmt.Println()
	fmt.Println("=== normal sync: two replicas, one is ahead ===")
	fmt.Println()

	replicaA := newReplica(0, 2)
	replicaB := newReplica(1, 2)

	replicaA.Write("hello", newVersionVector(2))
	fmt.Println("replica A writes 'hello', vector:", replicaA.versionVector, "dot:", replicaA.current.Dot)

	result := replicaB.Sync(replicaA)
	fmt.Println("replica B syncs from A:", result)
	fmt.Println("replica B value:", replicaB.current.Value, "vector:", replicaB.versionVector)

	replicaA.Write("hello-updated", copyOf(replicaA.versionVector))
	fmt.Println()
	fmt.Println("replica A updates to 'hello-updated', dot:", replicaA.current.Dot)

	result = replicaB.Sync(replicaA)
	fmt.Println("replica B syncs from A again:", result)
	fmt.Println("replica B value:", replicaB.current.Value, "vector:", replicaB.versionVector)

	// Sync again to show idempotency / already seen write
	result = replicaB.Sync(replicaA)
	fmt.Println("replica B syncs from A third time (idempotent):", result)
}
