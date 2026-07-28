// pulseright_cascade_route_test.go — PulseRight's cascade delta rules, the mirror of
// pulseleft_cascade_route_test.go.
//
// PulseRight (node 7) is a cascade TERMINUS with a sender whitelist:
//   - it ATTENDS to a delta triple only from a Time or SelectLeft
//     ("WindowAndInhibitRightGate") cascade neighbor — any other sender kind is dropped
//     outright in the moveMsgKindDeltaForward handler (no record, no relay);
//   - it NEVER relays a triple onward, from any sender and even as a direct drag
//     recipient (the guard at the top of forwardDelta).

package Wiring

import (
	"sync"
	"testing"
)

// TestForwardDeltaPulseRightNeverRelays: whatever the sender kind, a PulseRight forwards
// to nobody — it is the end of the cascade. The cascadeEdges here include the 7-9 double
// link, so this also pins the behavior the link's restoration depends on.
func TestForwardDeltaPulseRightNeverRelays(t *testing.T) {
	for _, sender := range []string{"4", "9"} {
		var mu sync.Mutex
		var got []string

		nm := &nodeMover{
			id:           "7",
			selfKind:     "PulseRight",
			cascadeEdges: []string{"4", "9"},
			cascadeKinds: map[string]string{"4": "Time", "9": "WindowAndInhibitRightGate"},
			sendMove: func(id string, msg moveMsg) {
				mu.Lock()
				got = append(got, id)
				mu.Unlock()
			},
		}

		nm.forwardDelta(nil, sender, 1, 2, 3)

		mu.Lock()
		if len(got) != 0 {
			t.Errorf("PulseRight relaying a delta from %q: want no forwards (terminus), got %v", sender, got)
		}
		mu.Unlock()
	}
}

// TestPulseRightAttendsOnlyTimeAndSelectLeft exercises the handler-level whitelist: a
// forward from Time or WindowAndInhibitRightGate is recorded (gotForwardMsg==1); one from
// any other kind leaves this node's forward state untouched.
func TestPulseRightAttendsOnlyTimeAndSelectLeft(t *testing.T) {
	cases := []struct {
		senderID   string
		senderKind string
		want       uint8
	}{
		{"4", "Time", 1},
		{"9", "WindowAndInhibitRightGate", 1},
		{"1", "Input", 0},
		{"8", "WindowAndInhibitLeftGate", 0},
	}
	for _, c := range cases {
		nm := &nodeMover{
			id:           "7",
			selfKind:     "PulseRight",
			cascadeEdges: []string{"4", "9"},
			cascadeKinds: map[string]string{"4": "Time", "9": "WindowAndInhibitRightGate",
				"1": "Input", "8": "WindowAndInhibitLeftGate"},
			sendMove: func(id string, msg moveMsg) {},
		}
		nm.handle(moveMsg{Kind: moveMsgKindDeltaForward, NodeID: "7", SenderID: c.senderID,
			DeltaA: 1, DeltaB: 2, DeltaC: 3})

		if nm.gotForwardMsg != c.want {
			t.Errorf("PulseRight delta from %s (%s): gotForwardMsg=%d, want %d",
				c.senderID, c.senderKind, nm.gotForwardMsg, c.want)
		}
	}
}
