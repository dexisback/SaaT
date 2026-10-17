package dvv

import "fmt"

// VersionVector tracks the causal history across replicas.
type VersionVector []int

func NewVersionVector(numberOfReplicas int) VersionVector {
	return make(VersionVector, numberOfReplicas)
}

func CopyOf(original VersionVector) VersionVector {
	copied := make(VersionVector, len(original))
	copy(copied, original)
	return copied
}

// Dot identifies one specific write event: which replica wrote it and at what counter value.
type Dot struct {
	ReplicaID int
	Counter   int
}

func (dot Dot) String() string {
	return fmt.Sprintf("(replica %d, counter %d)", dot.ReplicaID, dot.Counter)
}

// HasSeen checks whether a version vector includes a specific dot.
func (vector VersionVector) HasSeen(dot Dot) bool {
	if dot.ReplicaID < 0 || dot.ReplicaID >= len(vector) {
		return false
	}
	return vector[dot.ReplicaID] >= dot.Counter
}

// DottedValue stores the value itself, the Dot that created it, and the known context vector.
type DottedValue struct {
	Value         string
	Dot           Dot
	KnownVersions VersionVector
}

// Replica stores a key's current dotted value and the replica's version vector.
type Replica struct {
	ReplicaID     int
	Current       *DottedValue
	VersionVector VersionVector
}

func NewReplica(replicaID int, numberOfReplicas int) *Replica {
	return &Replica{
		ReplicaID:     replicaID,
		VersionVector: NewVersionVector(numberOfReplicas),
	}
}

// Read returns the current value and the version vector context.
func (replica *Replica) Read() (string, VersionVector) {
	if replica.Current == nil {
		return "", CopyOf(replica.VersionVector)
	}
	return replica.Current.Value, CopyOf(replica.VersionVector)
}

// Write stores a new value with a fresh Dot and client context.
func (replica *Replica) Write(newValue string, clientContext VersionVector) {
	replica.VersionVector[replica.ReplicaID]++
	currentCounter := replica.VersionVector[replica.ReplicaID]

	newDot := Dot{ReplicaID: replica.ReplicaID, Counter: currentCounter}
	replica.Current = &DottedValue{
		Value:         newValue,
		Dot:           newDot,
		KnownVersions: CopyOf(clientContext),
	}

	for slot := range replica.VersionVector {
		if clientContext[slot] > replica.VersionVector[slot] {
			replica.VersionVector[slot] = clientContext[slot]
		}
	}
}

// Sync merges another replica's dotted value into this one.
func (replica *Replica) Sync(other *Replica) string {
	if other.Current == nil {
		return "other replica has no value, nothing to sync"
	}

	incomingDot := other.Current.Dot
	incomingCounter := incomingDot.Counter
	ourKnowledge := replica.VersionVector[incomingDot.ReplicaID]

	if incomingCounter <= ourKnowledge {
		return fmt.Sprintf("already have dot %v, nothing to do", incomingDot)
	}

	replica.Current = &DottedValue{
		Value:         other.Current.Value,
		Dot:           other.Current.Dot,
		KnownVersions: CopyOf(other.Current.KnownVersions),
	}

	for slot := range replica.VersionVector {
		if other.VersionVector[slot] > replica.VersionVector[slot] {
			replica.VersionVector[slot] = other.VersionVector[slot]
		}
	}

	return fmt.Sprintf("accepted new value '%s' with dot %v", other.Current.Value, incomingDot)
}
