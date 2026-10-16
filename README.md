# SaaT

**SaaT** (*Saat*, Turkish for *"Time"*) is an exhaustive set of distributed systems clocks and causal ordering algorithms implemented from scratch in Go.

Implementations of every major distributed systems clock, each in a single self-contained Go file. No frameworks, no shared state between folders. Run any clock with `go run ./<folder>`.

---

## the progression

```
physical clocks     → skew breaks causality
lamport             → fixes causality, can't detect concurrency
vector              → detects concurrency, fixed node count
hlc                 → keeps causality and stays close to real time
truetime            → bounds uncertainty, waits it out before committing
version vectors     → same idea applied to replica conflict detection
dotted VV           → closes the silent-overwrite hole in version vectors
lww                 → skips conflict detection entirely, picks a winner
itc                 → drops the fixed node count, handles dynamic membership
matrix              → adds a second dimension: what does each node know others have seen
hybrid vector       → combines vector clock structure with HLC timestamps per slot
```

---

## clocks

- [`01-physical`](./01-physical) — wall clock with injectable skew; shows drift and non-monotonicity
- [`02-lamport`](./02-lamport) — single counter + node ID tie-break; guarantees cause < effect
- [`03-vector`](./03-vector) — per-node counter vector; detects true concurrency
- [`04-HLC`](./04-HLC) — hybrid logical clock; causality without losing proximity to real time
- [`05-truetime`](./05-truetime) — uncertainty interval + commit wait; Spanner's external consistency model
- [`06-version-vectors`](./06-version-vectors) — per-replica write tracking; detects conflicts across replicas
- [`07-dotted-version-vector`](./07-dotted-version-vector) — adds a dot per write; fixes silent data loss in version vectors
- [`08-LWW`](./08-LWW) — last write wins register; wall-clock and Lamport variants
- [`09-interval-tree-clocks`](./09-interval-tree-clocks) — interval tree clocks; fork/join identity without a fixed node count
- [`10-matrix`](./10-matrix) — N×N clock grid; tells you what every node knows every other node has seen
- [`11-hybrid-vector`](./11-hybrid-vector) — HLC timestamps per vector slot; concurrency detection with real time

---

## structure

Each folder has one `clock.go` (`package main`, runnable directly). Intentionally kept a single-file structure. Folders that benefit from real network behaviour have a `realtime/` or `realNTP/` subfolder that uses actual HTTP and, where relevant, live NTP queries.

```
/sim                     shared event harness (used by 01-physical)
/01-physical
    clock.go
    realNTP/             queries real NTP servers via github.com/beevik/ntp
/02-lamport
    clock.go
    realtime/            three HTTP nodes you drive from the keyboard
/03-vector
    clock.go
/04-HLC
    clock.go
    realtime/            real time.Now() with configurable per-node skew
/05-truetime
    clock.go
    realtime/            uncertainty window from real NTP round-trip times
/06-version-vectors
    clock.go
/07-dotted-version-vector
    clock.go
/08-LWW
    clock.go
/09-interval-tree-clocks
    clock.go
/10-matrix
    clock.go
/11-hybrid-vector
    clock.go
```

---

## run any clock

```bash
# Core self-contained simulations
go run ./03-vector
go run ./04-HLC
go run ./05-truetime
go run ./09-interval-tree-clocks

# Realtime network implementations
go run ./01-physical/realNTP
go run ./05-truetime/realtime
```

---

Built w love by dextertwts / amaanworks / amaan <3