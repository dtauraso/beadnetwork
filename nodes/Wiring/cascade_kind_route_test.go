// cascade_kind_route_test.go — kind-selective cascade delta-forward routing.
//
// forwardDelta normally floods a delta triple to every cascadeEdges neighbor except the
// sender. The one exception: a TimeStart node relaying a delta that arrived from a
// Pulse-kind cascade neighbor forwards ONLY to its Time-kind cascade neighbors. Every
// other node kind (and every non-Pulse sender) keeps the plain flood.

package Wiring

import (
	"sort"
	"sync"
	"testing"
)

func TestForwardDeltaTimeStartRoutesPulseOriginToTimeOnly(t *testing.T) {
	var mu sync.Mutex
	var got []string

	nm := &nodeMover{
		id:           "ts",
		selfKind:     "TimeStart",
		cascadeEdges: []string{"1", "4", "5"},
		cascadeKinds: map[string]string{"1": "Input", "4": "Time", "5": "Pulse"},
		sendMove: func(id string, msg moveMsg) {
			mu.Lock()
			got = append(got, id)
			mu.Unlock()
		},
	}

	nm.forwardDelta(nil, "5", 1, 2, 3)

	mu.Lock()
	defer mu.Unlock()
	if len(got) != 1 || got[0] != "4" {
		t.Fatalf("TimeStart relaying a Pulse-origin delta: want forward to [4] only, got %v", got)
	}
}

func TestForwardDeltaNonTimeStartFloodsAllExceptSender(t *testing.T) {
	var mu sync.Mutex
	var got []string

	nm := &nodeMover{
		id:           "other",
		selfKind:     "Foo",
		cascadeEdges: []string{"1", "4", "5"},
		cascadeKinds: map[string]string{"1": "Input", "4": "Time", "5": "Pulse"},
		sendMove: func(id string, msg moveMsg) {
			mu.Lock()
			got = append(got, id)
			mu.Unlock()
		},
	}

	nm.forwardDelta(nil, "5", 1, 2, 3)

	mu.Lock()
	defer mu.Unlock()
	sort.Strings(got)
	if len(got) != 2 || got[0] != "1" || got[1] != "4" {
		t.Fatalf("non-TimeStart mover: want flood to [1 4], got %v", got)
	}
}
