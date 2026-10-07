package sim

import (
	"fmt"
	"testing"
)

func TestPrintTrace(t *testing.T) {
	events := RunBasicScenario(69) //noice
	for _, e := range events {
		fmt.Printf("%v, %v, %v, %v\n", e.NodeID, e.Sequence, e.Type, e.MessageID)
	}
}
