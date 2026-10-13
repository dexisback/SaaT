#### demo.go:

- how a single perNode looks like:
0: []sim.Event{
        {NodeID: 0, Sequence: 0, Type: sim.Local, MsgID: 0},
        {NodeID: 0, Sequence: 1, Type: sim.Send,  MsgID: 100},
    },


- why do we structure perNode like that? Isolation: it separates the timeline of Node 0 from Node 1 and Node 2 so we can process each machine's events independently.
