package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
)

// clock will exactly the same as it was in clock.go
type Timestamp struct {
	Counter int
	NodeID  int
}

type Clock struct {
	lock    sync.Mutex
	nodeID  int
	counter int
}

func NewClock(nodeID int) *Clock {
	return &Clock{nodeID: nodeID, counter: 0}
}

func (clock *Clock) LocalEvent() Timestamp {
	clock.lock.Lock()
	defer clock.lock.Unlock()

	clock.counter++
	return Timestamp{Counter: clock.counter, NodeID: clock.nodeID}
}

func (clock *Clock) SendMessage() Timestamp {
	return clock.LocalEvent()
}

func (clock *Clock) ReceiveMessage(timestampInMessage Timestamp) Timestamp {
	clock.lock.Lock()
	defer clock.lock.Unlock()

	if timestampInMessage.Counter > clock.counter {
		clock.counter = timestampInMessage.Counter
	}
	clock.counter++
	return Timestamp{Counter: clock.counter, NodeID: clock.nodeID}
}

///    now, ---what travels over the network  ---

type Message struct {
	FromNode  int
	Timestamp Timestamp
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("nodeID must be a number like 0, 1 or 2")
		return
	}
	nodeID, err := strconv.Atoi(os.Args[1])
	if err != nil {
		fmt.Println("invalid nodeID:", err)
		return
	}

	clock := NewClock(nodeID)
	meraPort := 9000 + nodeID // node 0 listens to 9000, node 1 on 9001, etc

	// receiving side, runs whenever another node sends us something:
	http.HandleFunc("/message", func(w http.ResponseWriter, r *http.Request) {
		var message Message
		if err := json.NewDecoder(r.Body).Decode(&message); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		timestamp := clock.ReceiveMessage(message.Timestamp)
		fmt.Printf("[node %d] RECEIVED from node %d (their counter was %d) -> my counter is now %d\n",
			nodeID, message.FromNode, message.Timestamp.Counter, timestamp.Counter)
	})

	go func() {
		if err := http.ListenAndServe(fmt.Sprintf(":%d", meraPort), nil); err != nil {
			fmt.Println("server error:", err)
			os.Exit(1)
		}
	}()

	fmt.Printf("node %d is up on port %d. commands: local | send <node>\n", nodeID, meraPort)

	// typing side: read commands from the keyboard:
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		words := strings.Fields(scanner.Text())
		if len(words) == 0 {
			continue
		}

		switch words[0] {
		case "local":
			timestamp := clock.LocalEvent()
			fmt.Printf("[node %d] LOCAL event -> counter is now %d\n", nodeID, timestamp.Counter)

		case "send":
			if len(words) < 2 {
				fmt.Println("usage: send <node>")
				continue
			}
			targetNode, err := strconv.Atoi(words[1])
			if err != nil {
				fmt.Println("node must be a number")
				continue
			}

			timestamp := clock.SendMessage()
			body, _ := json.Marshal(Message{FromNode: nodeID, Timestamp: timestamp})
			url := fmt.Sprintf("http://localhost:%d/message", 9000+targetNode)
			if _, err := http.Post(url, "application/json", bytes.NewReader(body)); err != nil {
				fmt.Println("could not reach node", targetNode, "- is it running?")
				continue
			}
			fmt.Printf("[node %d] SENT to node %d -> counter is now %d\n", nodeID, targetNode, timestamp.Counter)

		default:
			fmt.Println("commands: local | send <node>")
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("input error:", err)
	}
}
