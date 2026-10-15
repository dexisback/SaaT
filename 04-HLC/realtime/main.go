//clock logic = same. difference = reads the real time itself (using time.Now()), and you pass a skew when you start each node. so we can make one machine's clock deliberately wrong
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
	"time"
)

// ---- the clock: same rules as 04-hlc/clock.go, but with REAL time ----

type Timestamp struct {
	WallTime int64
	Counter  int
}

func (timestamp Timestamp) String() string {
	return fmt.Sprintf("(%d, %d)", timestamp.WallTime, timestamp.Counter)
}

func biggerOf(first int64, second int64) int64 {
	if first > second {
		return first
	}
	return second
}

type Clock struct {
	lock             sync.Mutex
	skewMilliseconds int64 // how wrong this node's physical clock is, on purpose
	wallTime         int64
	counter          int
}

func NewClock(skewMilliseconds int64) *Clock {
	return &Clock{skewMilliseconds: skewMilliseconds}
}

// PhysicalTimeNow is the real time, plus this node's deliberate error.
func (clock *Clock) PhysicalTimeNow() int64 {
	return time.Now().UnixMilli() + clock.skewMilliseconds
}

func (clock *Clock) LocalEvent() Timestamp {
	clock.lock.Lock()
	defer clock.lock.Unlock()

	physicalTimeNow := clock.PhysicalTimeNow()
	previousWallTime := clock.wallTime
	clock.wallTime = biggerOf(previousWallTime, physicalTimeNow)

	if clock.wallTime == previousWallTime {
		clock.counter++
	} else {
		clock.counter = 0
	}
	return Timestamp{WallTime: clock.wallTime, Counter: clock.counter}
}

func (clock *Clock) SendMessage() Timestamp {
	return clock.LocalEvent()
}

func (clock *Clock) ReceiveMessage(timestampInMessage Timestamp) Timestamp {
	clock.lock.Lock()
	defer clock.lock.Unlock()

	physicalTimeNow := clock.PhysicalTimeNow()
	previousWallTime := clock.wallTime
	clock.wallTime = biggerOf(biggerOf(previousWallTime, timestampInMessage.WallTime), physicalTimeNow)

	sameAsMine := clock.wallTime == previousWallTime
	sameAsMessage := clock.wallTime == timestampInMessage.WallTime

	if sameAsMine && sameAsMessage {
		if timestampInMessage.Counter > clock.counter {
			clock.counter = timestampInMessage.Counter
		}
		clock.counter++
	} else if sameAsMine {
		clock.counter++
	} else if sameAsMessage {
		clock.counter = timestampInMessage.Counter + 1
	} else {
		clock.counter = 0
	}
	return Timestamp{WallTime: clock.wallTime, Counter: clock.counter}

}

//-----------what travels over the network------------:
//so we can see how far apart they are:

type Message struct {
	FromNode  int
	Timestamp Timestamp
}


func describe(clock *Clock, timestamp Timestamp) string {
	physicalTimeNow := clock.PhysicalTimeNow()
	return fmt.Sprintf("HLC %v | my physical clock says %d | HLC minus physical = %d ms", timestamp, physicalTimeNow, timestamp.WallTime-physicalTimeNow)
}

func main() {
	if len(os.Args) < 3 {
		fmt.Println("usage: go run ./04-HLC/realtime <nodeID> <skewMilliseconds>")
		return 
	}


	nodeID, err := strconv.Atoi(os.Args[1])
	if err != nil {
		fmt.Println("nodeID must be a number like 0/1/2")
		return 
	}
	skewMilliseconds, err := strconv.ParseInt(os.Args[2], 10, 64)
	if err != nil {
		fmt.Println("skew must be a number of milliseconds, like 0 or -5000")
		return 
	}
	clock := NewClock(skewMilliseconds)
	port := 9100 + nodeID //different from the lamport ports, so both can run 

	http.HandleFunc("/message", func(response http.ResponseWriter, request *http.Request) {
		var message Message
		err := json.NewDecoder(request.Body).Decode(&message)
		if err != nil {
			return
		}
		timestamp := clock.ReceiveMessage(message.Timestamp)
		fmt.Printf("[node %d] RECEIVED from node %d, they stamped %v\n          %s\n",
			nodeID, message.FromNode, message.Timestamp, describe(clock, timestamp))
	})
	go http.ListenAndServe(fmt.Sprintf(":%d", port), nil)

	fmt.Printf("node %d is up on port %d, clock skew %d ms. commands: local | send <node>\n",
		nodeID, port, skewMilliseconds)

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		words := strings.Fields(scanner.Text())
		if len(words) == 0 {
			continue
		}

		switch words[0] {
		case "local":
			timestamp := clock.LocalEvent()
			fmt.Printf("[node %d] LOCAL event\n          %s\n", nodeID, describe(clock, timestamp))

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
			url := fmt.Sprintf("http://localhost:%d/message", 9100+targetNode)
			_, err = http.Post(url, "application/json", bytes.NewReader(body))
			if err != nil {
				fmt.Println("could not reach node", targetNode, "- is it running?")
				continue
			}
			fmt.Printf("[node %d] SENT to node %d\n          %s\n", nodeID, targetNode, describe(clock, timestamp))

		default:
			fmt.Println("commands: local | send <node>")
		}
	}
	
}



