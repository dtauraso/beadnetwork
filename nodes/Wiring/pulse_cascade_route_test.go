// pulse_cascade_route_test.go — the Pulse kind's cascade delta rule.
//
// A Pulse (node 5) IGNORES a delta triple arriving from a TimeStart-kind cascade
// neighbor: no record, no relay. From every other sender kind it keeps the plain flood
// to all cascade neighbors except the sender. This is the Pulse kind only — PulseLeft
// and PulseRight are separate kinds with their own termini rules.

package Wiring

import (
	"sort"
	"sync"
	"testing"
)

// TestPulseIgnoresTimeStartOriginDelta: a delta from a TimeStart sender is dropped
// outright — gotForwardMsg stays 0 (not recorded) and nothing is relayed.
func TestPulseIgnoresTimeStartOriginDelta(t *testing.T) {
	var mu sync.Mutex
	var got []string

	nm := &nodeMover{
		id:           "5",
		selfKind:     "Pulse",
		cascadeEdges: []string{"2", "9", "8"},
		cascadeKinds: map[string]string{"2": "TimeStart", "9": "WindowAndInhibitRightGate",
			"8": "WindowAndInhibitLeftGate"},
		sendMove: func(id string, msg moveMsg) {
			mu.Lock()
			got = append(got, id)
			mu.Unlock()
		},
	}
	nm.forwardOnce = func(exceptID string, dA, dB, dC int32) {
		nm.forwardDelta(nil, exceptID, dA, dB, dC)
	}

	nm.handle(moveMsg{Kind: moveMsgKindDeltaForward, NodeID: "5", SenderID: "2",
		DeltaA: 1, DeltaB: 2, DeltaC: 3})

	mu.Lock()
	defer mu.Unlock()
	if nm.gotForwardMsg != 0 {
		t.Errorf("Pulse delta from 2 (TimeStart): gotForwardMsg=%d, want 0 (ignored)", nm.gotForwardMsg)
	}
	if len(got) != 0 {
		t.Errorf("Pulse delta from 2 (TimeStart): relayed to %v, want none (ignored)", got)
	}
}

// TestPulseFloodsNonTimeStartOrigin: the ignore rule is scoped to TimeStart senders —
// a delta from a gate neighbor is recorded AND flooded to the other cascade neighbors.
func TestPulseFloodsNonTimeStartOrigin(t *testing.T) {
	var mu sync.Mutex
	var got []string

	nm := &nodeMover{
		id:           "5",
		selfKind:     "Pulse",
		cascadeEdges: []string{"2", "9", "8"},
		cascadeKinds: map[string]string{"2": "TimeStart", "9": "WindowAndInhibitRightGate",
			"8": "WindowAndInhibitLeftGate"},
		sendMove: func(id string, msg moveMsg) {
			mu.Lock()
			got = append(got, id)
			mu.Unlock()
		},
	}
	nm.forwardOnce = func(exceptID string, dA, dB, dC int32) {
		nm.forwardDelta(nil, exceptID, dA, dB, dC)
	}

	nm.handle(moveMsg{Kind: moveMsgKindDeltaForward, NodeID: "5", SenderID: "9",
		DeltaA: 1, DeltaB: 2, DeltaC: 3})

	mu.Lock()
	defer mu.Unlock()
	if nm.gotForwardMsg != 1 {
		t.Errorf("Pulse delta from 9 (gate): gotForwardMsg=%d, want 1 (attended)", nm.gotForwardMsg)
	}
	sort.Strings(got)
	if len(got) != 2 || got[0] != "2" || got[1] != "8" {
		t.Errorf("Pulse delta from 9 (gate): relayed to %v, want [2 8]", got)
	}
}

// TestPulseLeftRightUnaffectedByPulseRule pins that the rule is keyed on the Pulse kind
// exactly: PulseLeft and PulseRight are distinct kinds and keep their own behavior (both
// are termini that never relay, and each has its own sender whitelist).
func TestPulseLeftRightUnaffectedByPulseRule(t *testing.T) {
	for _, kind := range []string{"PulseLeft", "PulseRight"} {
		var mu sync.Mutex
		var got []string
		nm := &nodeMover{
			id:           "x",
			selfKind:     kind,
			cascadeEdges: []string{"2", "9"},
			cascadeKinds: map[string]string{"2": "TimeStart", "9": "WindowAndInhibitRightGate"},
			sendMove: func(id string, msg moveMsg) {
				mu.Lock()
				got = append(got, id)
				mu.Unlock()
			},
		}
		nm.forwardDelta(nil, "9", 1, 2, 3)
		mu.Lock()
		if len(got) != 0 {
			t.Errorf("%s: relayed to %v, want none (terminus)", kind, got)
		}
		mu.Unlock()
	}
}
