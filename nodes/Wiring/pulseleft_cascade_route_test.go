// pulseleft_cascade_route_test.go — PulseLeft's cascade delta rules.
//
// PulseLeft (node 3) is a cascade TERMINUS with a sender whitelist:
//   - it ATTENDS to a delta triple only from an Input or SelectRight cascade
//     neighbor — any other sender kind is dropped outright in the
//     moveMsgKindDeltaForward handler (no record, no relay);
//   - it NEVER relays a triple onward, from any sender and even as a direct drag
//     recipient (the guard at the top of forwardDelta).

package Wiring

import (
	"sync"
	"testing"
)

// TestForwardDeltaPulseLeftNeverRelays: whatever the sender kind, a PulseLeft forwards to
// nobody — it is the end of the cascade.
func TestForwardDeltaPulseLeftNeverRelays(t *testing.T) {
	for _, sender := range []string{"1", "8"} {
		var mu sync.Mutex
		var got []string

		nm := &nodeMover{
			id:           "3",
			selfKind:     "PulseLeft",
			cascadeEdges: []string{"1", "8"},
			cascadeKinds: map[string]string{"1": "Input", "8": "SelectRight"},
			sendMove: func(id string, msg moveMsg) {
				mu.Lock()
				got = append(got, id)
				mu.Unlock()
			},
		}

		nm.forwardDelta(nil, sender, 1, 2, 3)

		mu.Lock()
		if len(got) != 0 {
			t.Errorf("PulseLeft relaying a delta from %q: want no forwards (terminus), got %v", sender, got)
		}
		mu.Unlock()
	}
}

// TestPulseLeftAttendsOnlyInputAndSelectRight exercises the handler-level whitelist: a
// forward from Input or SelectRight is recorded (gotForwardMsg==1); one from
// any other kind leaves this node's forward state untouched.
func TestPulseLeftAttendsOnlyInputAndSelectRight(t *testing.T) {
	cases := []struct {
		senderID   string
		senderKind string
		want       uint8
	}{
		{"1", "Input", 1},
		{"8", "SelectRight", 1},
		{"4", "Time", 0},
		{"5", "Pulse", 0},
	}
	for _, c := range cases {
		nm := &nodeMover{
			id:           "3",
			selfKind:     "PulseLeft",
			cascadeEdges: []string{"1", "8"},
			cascadeKinds: map[string]string{"1": "Input", "8": "SelectRight", "4": "Time", "5": "Pulse"},
			sendMove:     func(id string, msg moveMsg) {},
		}
		nm.handle(moveMsg{Kind: moveMsgKindDeltaForward, NodeID: "3", SenderID: c.senderID,
			DeltaA: 1, DeltaB: 2, DeltaC: 3})

		if nm.gotForwardMsg != c.want {
			t.Errorf("PulseLeft delta from %s (%s): gotForwardMsg=%d, want %d",
				c.senderID, c.senderKind, nm.gotForwardMsg, c.want)
		}
	}
}
