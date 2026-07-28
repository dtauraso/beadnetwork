// cascade_kind_route_test.go — kind-directed cascade delta-forward routing.
//
// forwardDelta normally floods a delta triple to every cascadeEdges neighbor except the
// sender. A TimeStart node instead ROUTES by the sender's kind: Pulse->Time, Time->Pulse,
// and pays attention only to Pulse/Time/Input senders — a delta from any other kind is
// dropped entirely. Every non-TimeStart node keeps the plain flood.

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

func TestForwardDeltaTimeStartRoutesTimeOriginToPulseOnly(t *testing.T) {
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

	nm.forwardDelta(nil, "4", 1, 2, 3)

	mu.Lock()
	defer mu.Unlock()
	if len(got) != 1 || got[0] != "5" {
		t.Fatalf("TimeStart relaying a Time-origin delta: want forward to [5] only, got %v", got)
	}
}

// TestForwardDeltaTimeStartIgnoresNonPTIOrigin: a TimeStart drops a delta arriving from a
// neighbor whose kind is not Pulse/Time/Input (here a gate) — it forwards to nobody.
func TestForwardDeltaTimeStartIgnoresNonPTIOrigin(t *testing.T) {
	var mu sync.Mutex
	var got []string

	nm := &nodeMover{
		id:           "ts",
		selfKind:     "TimeStart",
		cascadeEdges: []string{"4", "5", "8"},
		cascadeKinds: map[string]string{"4": "Time", "5": "Pulse", "8": "SelectRight"},
		sendMove: func(id string, msg moveMsg) {
			mu.Lock()
			got = append(got, id)
			mu.Unlock()
		},
	}

	nm.forwardDelta(nil, "8", 1, 2, 3)

	mu.Lock()
	defer mu.Unlock()
	if len(got) != 0 {
		t.Fatalf("TimeStart relaying a gate-origin delta: want no forwards (dropped), got %v", got)
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
